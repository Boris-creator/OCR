package media

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/telebot.v4"
	"io"
	"log/slog"
	"os"
	"path"
	"tele/internal/api"
	"tele/internal/db/query"
	"tele/internal/mistral"
	"tele/internal/s3"
)

type Handler struct {
	api.Handler
	mc *mistral.Client
	s3 *s3.Storage
	db *pgxpool.Pool
}

func New(b *telebot.Bot, mc *mistral.Client, s3 *s3.Storage, db *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{
		*api.New(b, logger),
		mc, s3, db,
	}
}

func (handler *Handler) Handle(ctx telebot.Context) error {
	const errPrefix = "media.Handle"

	c := context.TODO()

	doc := ctx.Message().Photo
	if doc == nil {
		return nil
	}

	text, err := handler.getImageOCR(c, doc, ctx.Chat().ID, func(err error, msg string) error {
		return fmt.Errorf("%s: %s: %v", errPrefix, msg, err)
	})
	if err != nil {
		return handler.InternalErrorResponse(ctx, fmt.Errorf("%s: %w", errPrefix, err))
	}
	if len(text) == 0 {
		return handler.noTextFoundResponse(ctx)
	}

	return handler.successResponse(ctx, text)
}

func (handler *Handler) successResponse(ctx telebot.Context, text string) error {
	const maxMsgTextLen = 1 << 12

	symbols := []rune(text)
	handler.Logger.Debug(string(symbols))

	return ctx.Reply(string(symbols[:min(len(symbols), maxMsgTextLen)]))
}

func (handler *Handler) noTextFoundResponse(ctx telebot.Context) error {
	return ctx.Reply("Text not found")
}

func (handler *Handler) getImageOCR(
	ctx context.Context,
	doc *telebot.Photo,
	chatId int64,
	wrapError func(error, string) error,
) (string, error) {
	var res string

	bot := handler.Bot

	userFile, err := bot.FileByID(doc.FileID)
	if err != nil {
		return res, fmt.Errorf("bot.FileByID %s: %w", doc.FileID, err)
	}

	rc, err := bot.File(&userFile)
	if err != nil {
		return res, fmt.Errorf("bot.File: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	fileBytes, err := io.ReadAll(rc)
	if err != nil {
		return res, fmt.Errorf("read file: %s", err)
	}

	tx, err := handler.db.Begin(ctx)
	if err != nil {
		return res, fmt.Errorf("db.Begin: %w", err)
	}
	defer func(err error) {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}(err)

	queryHandler := query.New(tx)

	hash := getFileCheckSum(fileBytes)

	document, err := queryHandler.GetDocumentByHash(ctx, query.GetDocumentByHashParams{
		Hash:   pgtype.UUID{Bytes: hash, Valid: true},
		ChatID: chatId,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		handler.Logger.Error(wrapError(err, "query.GetDocumentByHash").Error())
	}
	if err == nil {
		var ocr mistral.OCRResponse
		_ = json.Unmarshal(document.Ocr, &ocr)
		text, _ := getOCRText(ocr)

		return text, nil
	}

	ocr, err := handler.mc.ProcessFile(bytes.NewReader(fileBytes), doc.UniqueID, mistral.ImageUrl)
	if err != nil {
		return res, fmt.Errorf("mistral.ProcessFile: %w", err)
	}

	ocrData, _ := json.Marshal(ocr)
	newDocumentId, savingErr := queryHandler.CreateDocument(ctx, query.CreateDocumentParams{
		FileID: doc.FileID,
		ChatID: chatId,
		Hash:   pgtype.UUID{Bytes: hash, Valid: true},
		Ocr:    ocrData,
	})

	if savingErr != nil {
		handler.Logger.Error(savingErr.Error())
	}

	if savingErr == nil {
		go func() {
			file, err := os.CreateTemp("", "*")

			defer func(err error) {
				if err != nil {
					handler.Logger.Error(err.Error())
				}
			}(err)

			if err != nil {
				err = wrapError(err, "os.CreateTemp")
				return
			}
			defer func() {
				_ = file.Close()
				_ = os.Remove(file.Name())
			}()

			_, _ = io.Copy(file, bytes.NewReader(fileBytes))

			err = handler.s3.UploadFromLocal(ctx, file.Name(), fmt.Sprintf("%d%s", newDocumentId, path.Ext(userFile.FilePath)))
			if err != nil {
				err = wrapError(err, "s3.UploadFromLocal")
				return
			}

			err = tx.Commit(ctx)
			if err != nil {
				err = wrapError(err, "tx.Commit")
			}
		}()
	}

	res, _ = getOCRText(*ocr)

	return res, nil
}

func getFileCheckSum(file []byte) [16]byte {
	return md5.Sum(file)
}

func getOCRText(ocr mistral.OCRResponse) (string, bool) {
	if len(ocr.Pages) == 0 {
		return "", false
	}

	return ocr.Pages[0].Markdown, true
}

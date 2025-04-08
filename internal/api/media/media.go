package media

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"tele/internal/api"

	"gopkg.in/telebot.v4"
)

type Handler struct {
	api.Handler
	ocr imageTextRecognizer
}

func New(b *telebot.Bot, logger *slog.Logger, ocr imageTextRecognizer) *Handler {
	return &Handler{
		*api.New(b, logger),
		ocr,
	}
}

func (handler *Handler) Handle(tctx telebot.Context) error {
	const errPrefix = "media.Handle"

	ctx := context.TODO()

	imageFile, closeImageFile, err := handler.getImage(tctx)
	if err != nil {
		return handler.InternalErrorResponse(tctx, fmt.Errorf("%s: %w", errPrefix, err))
	}

	if imageFile == nil {
		return nil
	}

	defer closeImageFile()

	text, err := handler.ocr.GetImageOCR(ctx, file{*imageFile}, tctx.Chat().ID)
	if err != nil {
		return handler.InternalErrorResponse(tctx, fmt.Errorf("%s: imageTextRecognizer.GetImageOCR: %w", errPrefix, err))
	}

	if len(text) == 0 {
		return handler.noTextFoundResponse(tctx)
	}

	return handler.successResponse(tctx, text)
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

func (handler *Handler) getImage(c telebot.Context) (file *telebot.File, closeFile func(), err error) {
	var fileID string
	var imageFound bool

	photo := c.Message().Photo
	if photo != nil {
		imageFound = true
		fileID = photo.FileID
	} else {
		doc := c.Message().Document
		if doc != nil && strings.HasPrefix(doc.MIME, "image") {
			imageFound = true
			fileID = doc.FileID
		}
	}

	if !imageFound {
		return nil, nil, nil
	}

	bot := handler.Bot

	userFile, err := bot.FileByID(fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("bot.FileByID %s: %w", fileID, err)
	}

	frc, err := bot.File(&userFile)
	if err != nil {
		return nil, nil, fmt.Errorf("bot.File: %w", err)
	}

	userFile.FileReader = frc

	return &userFile, func() {
		_ = frc.Close()
	}, nil
}

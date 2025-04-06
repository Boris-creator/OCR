package media

import (
	"context"
	"fmt"
	"gopkg.in/telebot.v4"
	"log/slog"
	"tele/internal/api"
	"tele/internal/usecase/ocr"
)

type imageTextRecognizer interface {
	GetImageOCR(ctx context.Context, file ocr.File, chatId int64) (string, error)
}

type file struct {
	telebot.File
}

func (f file) Read(p []byte) (n int, err error) {
	return f.FileReader.Read(p)
}
func (f file) Id() string {
	return f.FileID
}
func (f file) Path() string {
	return f.FilePath
}

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

func (handler *Handler) Handle(ctx telebot.Context) error {
	const errPrefix = "media.Handle"

	c := context.TODO()

	imageFile, closeImageFile, err := handler.getImage(ctx)
	if err != nil {
		return handler.InternalErrorResponse(ctx, fmt.Errorf("%s: %w", errPrefix, err))
	}
	defer closeImageFile()

	text, err := handler.ocr.GetImageOCR(c, file{*imageFile}, ctx.Chat().ID)
	if err != nil {
		return handler.InternalErrorResponse(ctx, fmt.Errorf("%s: imageTextRecognizer.GetImageOCR: %w", errPrefix, err))
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

func (handler *Handler) getImage(c telebot.Context) (file *telebot.File, close func(), err error) {
	doc := c.Message().Photo
	if doc == nil {
		return nil, nil, nil
	}
	bot := handler.Bot

	userFile, err := bot.FileByID(doc.FileID)
	if err != nil {
		return nil, nil, fmt.Errorf("bot.FileByID %s: %w", doc.FileID, err)
	}

	rc, err := bot.File(&userFile)
	if err != nil {
		return nil, nil, fmt.Errorf("bot.File: %w", err)
	}

	userFile.FileReader = rc

	return &userFile, func() {
		_ = rc.Close()
	}, nil
}

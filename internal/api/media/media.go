package media

import (
	"bytes"
	"context"
	"fmt"
	"gopkg.in/telebot.v4"
	"io"
	"log"
	"os"
	"path"
	"tele/internal/mistral"
	"tele/internal/s3"
)

type Handler struct {
	mc mistral.Client
	s3 s3.Storage
}

func New(mc mistral.Client, s3 s3.Storage) *Handler {
	return &Handler{mc, s3}
}

func (handler Handler) Handle(ctx telebot.Context) error {
	const errPrefix = "bot.Handle"
	const maxMsgTextLen = 1 << 12

	bot := ctx.Bot()
	doc := ctx.Message().Photo
	if doc == nil {
		return nil
	}

	userFile, err := bot.FileByID(doc.FileID)
	if err != nil {
		return fmt.Errorf("%s: bot.FileByID %s: %w", errPrefix, doc.FileID, err)
	}

	rc, err := bot.File(&userFile)
	if err != nil {
		return fmt.Errorf("%s: bot.File: %w", errPrefix, err)
	}
	defer func() {
		_ = rc.Close()
	}()

	fileBytes, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("%s: read file: %s", errPrefix, err)
	}

	go func() {
		file, err := os.CreateTemp("", "*")
		if err != nil {
			log.Printf("%s: os.CreateTemp: %v\n", errPrefix, err)
			return
		}
		defer func() {
			_ = file.Close()
			_ = os.Remove(file.Name())
		}()

		_, _ = io.Copy(file, bytes.NewReader(fileBytes))

		err = handler.s3.UploadFromLocal(context.Background(), file.Name(), fmt.Sprintf("%s%s", doc.FileID, path.Ext(userFile.FilePath)))
		if err != nil {
			log.Printf("%s: s3.UploadFromLocal: %s\n", errPrefix, err.Error())
		}
	}()

	ocr, err := handler.mc.ProcessFile(bytes.NewReader(fileBytes), doc.UniqueID, mistral.ImageUrl)
	if err != nil {
		return handler.internalErrorResponse(ctx)
	}
	if len(ocr.Pages) == 0 {
		return handler.noTextFoundResponse(ctx)
	}

	res := ocr.Pages[0].Markdown
	text := []rune(res)
	log.Println(string(text))

	return ctx.Reply(string(text[:min(len(text), maxMsgTextLen)]))
}

func (handler Handler) noTextFoundResponse(ctx telebot.Context) error {
	return ctx.Reply("Text not found")
}

func (handler Handler) internalErrorResponse(ctx telebot.Context) error {
	return ctx.Reply("Internal error")
}

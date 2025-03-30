package media

import (
	"fmt"
	"gopkg.in/telebot.v4"
	"log"
	"os"
	"tele/internal/mistral"
)

type Handler struct {
	mc mistral.Client
}

func New(mc mistral.Client) *Handler {
	return &Handler{mc}
}

func (handler Handler) Handle(ctx telebot.Context) error {
	const errPrefix = "bot.Handle"
	const maxMsgTextLen = 1 << 12

	bot := ctx.Bot()
	doc := ctx.Message().Photo
	if doc == nil {
		return nil
	}
	file, _ := os.CreateTemp("", "*")
	defer func() {
		_ = os.Remove(file.Name())
	}()

	userFile, err := bot.FileByID(doc.FileID)
	if err != nil {
		return fmt.Errorf("%s: bot.FileByID %s: %w", errPrefix, doc.FileID, err)
	}

	rc, err := bot.File(&userFile)
	if err != nil {
		return fmt.Errorf("%s: bot.File: %w", errPrefix, err)
	}

	ocr, _ := handler.mc.ProcessFile(rc, doc.UniqueID, mistral.ImageUrl)
	if len(ocr.Pages) == 0 {
		return ctx.Send("Text not found")
	}

	res := ocr.Pages[0].Markdown
	text := []rune(res)
	log.Println(string(text))

	return ctx.Reply(string(text[:min(len(text), maxMsgTextLen)]))
}

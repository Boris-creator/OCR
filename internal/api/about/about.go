package about

import (
	"fmt"
	"log/slog"
	"tele/internal/api"

	"gopkg.in/telebot.v4"
)

type aboutService interface {
	GetSourceCodeURL() (string, error)
}

type Handler struct {
	api.Handler
	aboutService aboutService
}

func New(bot *telebot.Bot, aboutService aboutService, logger *slog.Logger) *Handler {
	return &Handler{
		*api.New(bot, logger),
		aboutService,
	}
}

func (handler *Handler) Handle(ctx telebot.Context) error {
	const errPrefix = "about.Handle"

	out, err := handler.aboutService.GetSourceCodeURL()
	if err != nil {
		return handler.InternalErrorResponse(ctx, fmt.Errorf("%s: %w", errPrefix, err))
	}

	return ctx.Send(fmt.Sprintf("See my source code at %s.", out))
}

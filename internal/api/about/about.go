package about

import (
	"fmt"
	"gopkg.in/telebot.v4"
	"log/slog"
	"tele/internal/api"
)

type aboutService interface {
	GetSourceCodeUrl() (string, error)
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

	out, err := handler.aboutService.GetSourceCodeUrl()
	if err != nil {
		return handler.InternalErrorResponse(ctx, fmt.Errorf("%s: %w", errPrefix, err))
	}

	return ctx.Send(fmt.Sprintf("See my source code at %s", out))
}

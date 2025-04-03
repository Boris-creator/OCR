package about

import (
	"fmt"
	"gopkg.in/telebot.v4"
	"log/slog"
	"os/exec"
	"tele/internal/api"
)

type Handler struct {
	api.Handler
}

func New(bot *telebot.Bot, logger *slog.Logger) *Handler {
	return &Handler{
		*api.New(bot, logger),
	}
}

func (handler *Handler) Handle(ctx telebot.Context) error {
	const errPrefix = "about.Handle"

	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return handler.InternalErrorResponse(ctx, fmt.Errorf("%s: command.Run %s: %s; %w", errPrefix, cmd.String(), out, err))
	}

	return ctx.Send(fmt.Sprintf("See my source code at %s", string(out)))
}

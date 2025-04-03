package api

import (
	"gopkg.in/telebot.v4"
	"log/slog"
)

type Handler struct {
	Bot    *telebot.Bot
	Logger *slog.Logger
}

func New(bot *telebot.Bot, logger *slog.Logger) *Handler {
	return &Handler{bot, logger}
}

func (handler *Handler) InternalErrorResponse(ctx telebot.Context, err error) error {
	handler.Logger.Error(err.Error())
	return ctx.Reply("Internal error")
}

package middleware

import (
	"context"
	"fmt"
	"gopkg.in/telebot.v4"
	"log/slog"
)

type chatRepository interface {
	CreateOrUpdateChat(ctx context.Context, int65 int64) error
}

type Activity struct {
	repo   chatRepository
	logger *slog.Logger
}

func NewActivityMiddleware(chatRepo chatRepository, logger *slog.Logger) *Activity {
	return &Activity{chatRepo, logger}
}

func (mw Activity) RegisterOrRecordRequest(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatId := c.Chat().ID

		err := mw.repo.CreateOrUpdateChat(context.Background(), chatId)
		if err != nil {
			mw.logger.Warn(fmt.Sprintf("middleware.RegisterOrRecordRequest: %v", err))
		}

		return next(c)
	}
}

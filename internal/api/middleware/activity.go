package middleware

import (
	"context"
	"fmt"
	"log/slog"

	"gopkg.in/telebot.v4"
)

type chatRepository interface {
	CreateOrUpdateChat(ctx context.Context, chatID int64) error
}

type Activity struct {
	repo   chatRepository
	logger *slog.Logger
}

func NewActivityMiddleware(chatRepo chatRepository, logger *slog.Logger) *Activity {
	return &Activity{chatRepo, logger}
}

func (mw Activity) RegisterOrRecordRequest(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(tctx telebot.Context) error {
		chatID := tctx.Chat().ID

		err := mw.repo.CreateOrUpdateChat(context.Background(), chatID)
		if err != nil {
			mw.logger.Warn(fmt.Sprintf("middleware.RegisterOrRecordRequest: %v", err))
		}

		return next(tctx)
	}
}

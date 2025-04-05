package middleware

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/telebot.v4"
	"log/slog"
	"tele/internal/db/query"
)

type Activity struct {
	baseMiddleware
}

func NewActivityMiddleware(db *pgxpool.Pool, logger *slog.Logger) *Activity {
	return &Activity{
		baseMiddleware{db, logger},
	}
}

func (mw Activity) RegisterOrRecordRequest(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatId := c.Chat().ID
		queryHandler := query.New(mw.db)

		err := queryHandler.CreateOrUpdateChat(context.Background(), chatId)
		if err != nil {
			mw.logger.Warn(fmt.Sprintf("middleware.RegisterOrRecordRequest: %v", err))
		}

		return next(c)
	}
}

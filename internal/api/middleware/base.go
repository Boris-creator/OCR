package middleware

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type baseMiddleware struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

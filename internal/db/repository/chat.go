package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	baseRepository
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{
		*newRepository(db),
	}
}

func (repo ChatRepository) CreateOrUpdateChat(ctx context.Context, chatID int64) error {
	err := repo.queries.CreateOrUpdateChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("repo.CreateOrUpdateChat: %w", err)
	}

	return nil
}

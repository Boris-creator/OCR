package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"tele/internal/db/query"
	"tele/internal/domain"
)

type DocumentRepository struct {
	baseRepository
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{
		*newRepository(db),
	}
}

func (repo DocumentRepository) GetDocumentByHash(ctx context.Context, hash [16]byte, chatId int64) (*domain.Document, bool, error) {
	document, err := repo.queries.GetDocumentByHash(ctx, query.GetDocumentByHashParams{
		Hash:   pgtype.UUID{Bytes: hash, Valid: true},
		ChatID: chatId,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("DocumentRepository.GetDocumentByHash: %w", err)
	}

	doc := domain.Document{
		Id:  document.ID,
		Ocr: document.Ocr,
	}

	return &doc, true, nil
}

func (repo DocumentRepository) CreateDocument(
	ctx context.Context,
	document interface {
		Params() (fileId string, chatId int64, hash [16]byte, ocr []byte)
	}) (createdDocumentId int64, err error) {
	fileId, chatId, hash, ocr := document.Params()
	id, err := repo.queries.CreateDocument(ctx, query.CreateDocumentParams{
		FileID: fileId,
		ChatID: chatId,
		Hash:   pgtype.UUID{Bytes: hash, Valid: true},
		Ocr:    ocr,
	})
	if err != nil {
		return 0, fmt.Errorf("DocumentRepository.CreateDocument: %w", err)
	}

	return id, nil
}

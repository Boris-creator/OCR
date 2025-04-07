package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"tele/internal/db/query"
)

type baseRepository struct {
	db      *pgxpool.Pool
	tx      *pgx.Tx
	queries *query.Queries
}

func newRepository(db *pgxpool.Pool) *baseRepository {
	queries := query.New(db)
	return &baseRepository{
		db:      db,
		queries: queries,
	}
}

var OutOfTransactionError = errors.New("out of transaction")

//TODO: change WithTx signature

func (repo *baseRepository) WithTx(ctx context.Context) (repoWithTx *baseRepository, err error) {
	tx, err := repo.beginTx(ctx)
	if err != nil {
		return nil, err
	}

	repoWithTx = &baseRepository{
		db:      repo.db,
		tx:      &tx,
		queries: query.New(tx),
	}

	return repoWithTx, nil
}

func (repo *baseRepository) BeginTx(ctx context.Context) error {
	tx, err := repo.beginTx(ctx)
	if err != nil {
		return err
	}

	repo.tx = &tx
	repo.queries = query.New(tx)

	return nil
}

func (repo *baseRepository) Commit(ctx context.Context) error {
	if repo.tx == nil {
		return OutOfTransactionError
	}
	err := (*repo.tx).Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	repo.tx = nil
	repo.queries = query.New(repo.db)

	return nil
}

func (repo *baseRepository) Rollback(ctx context.Context) error {
	if repo.tx == nil {
		return OutOfTransactionError
	}
	err := (*repo.tx).Rollback(ctx)
	if err != nil {
		return fmt.Errorf("tx.Rollback: %w", err)
	}

	repo.tx = nil
	repo.queries = query.New(repo.db)

	return nil
}

func (repo *baseRepository) beginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("baseRepository.BeginTx: %w", err)
	}

	return tx, nil
}

package ocr

import (
	"context"
	"io"
	"tele/internal/domain"
)

type File interface {
	io.Reader
	Id() string
	Path() string
}

type documentRepository interface {
	GetDocumentByHash(ctx context.Context, hash [16]byte, chatId int64) (*domain.Document, bool, error)
	CreateDocument(ctx context.Context, document interface {
		Params() (fileId string, chatId int64, hash [16]byte, ocr []byte)
	}) (int64, error)
	BeginTx(context.Context) error
	Commit(context.Context) error
	Rollback(context.Context) error
}

type documentParams struct {
	fileId string
	chatId int64
	hash   [16]byte
	ocr    []byte
}

func (d documentParams) Params() (fileId string, chatId int64, hash [16]byte, ocr []byte) {
	return d.fileId, d.chatId, d.hash, d.ocr
}

package ocr

import (
	"context"
	"io"
	"tele/internal/domain"
)

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

type fileStorage interface {
	UploadFromLocal(ctx context.Context, localFilePath string, destination string) error
}

type ocrService[R interface{ Text() string }] interface {
	GetImageOCR(file io.Reader, fileName string) (R, error)
}

type ocrResult interface {
	Text() string
}

package media

import (
	"context"
	"io"

	"gopkg.in/telebot.v4"
)

type imageTextRecognizer interface {
	GetImageOCR(
		ctx context.Context,
		file interface {
			io.Reader
			ID() string
			Path() string
		},
		chatID int64,
	) (string, error)
}

type file struct {
	telebot.File
}

func (f file) Read(p []byte) (int, error) {
	return f.FileReader.Read(p)
}
func (f file) ID() string {
	return f.FileID
}
func (f file) Path() string {
	return f.FilePath
}

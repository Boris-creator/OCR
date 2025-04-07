package media

import (
	"context"
	"gopkg.in/telebot.v4"
	"io"
)

type imageTextRecognizer interface {
	GetImageOCR(
		ctx context.Context,
		file interface {
			io.Reader
			Id() string
			Path() string
		},
		chatId int64,
	) (string, error)
}

type file struct {
	telebot.File
}

func (f file) Read(p []byte) (n int, err error) {
	return f.FileReader.Read(p)
}
func (f file) Id() string {
	return f.FileID
}
func (f file) Path() string {
	return f.FilePath
}

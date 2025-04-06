package ocr

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"tele/internal/mistral"
	"tele/internal/s3"
)

type ImageTextRecognizer struct {
	mc     *mistral.Client
	s3     *s3.Storage
	repo   documentRepository
	logger *slog.Logger
}

func New(mc *mistral.Client, s3 *s3.Storage, repo documentRepository, logger *slog.Logger) *ImageTextRecognizer {
	return &ImageTextRecognizer{mc, s3, repo, logger}
}

func (recognizer ImageTextRecognizer) GetImageOCR(
	ctx context.Context,
	userFile File,
	chatId int64,
) (string, error) {
	var res string

	wrapError := func(err error, msg string) error {
		return fmt.Errorf("%s: %s: %v", "ImageTextRecognizer.getImageOCR: ", msg, err)
	}

	fileBytes, err := io.ReadAll(userFile)
	if err != nil {
		return res, wrapError(err, "read file")
	}

	rep := recognizer.repo
	err = rep.BeginTx(ctx)
	if err != nil {
		return res, wrapError(err, "db.Begin")
	}
	defer func(err error) {
		if err != nil {
			_ = rep.Rollback(ctx)
		}
	}(err)

	hash := getFileCheckSum(fileBytes)

	document, ok, err := rep.GetDocumentByHash(ctx, hash, chatId)
	if err != nil {
		recognizer.logger.Error(wrapError(err, "query.GetDocumentByHash").Error())
	}
	if ok {
		var ocr mistral.OCRResponse
		_ = json.Unmarshal(document.Ocr, &ocr)
		text, _ := getOCRText(ocr)

		return text, nil
	}

	fileId := userFile.Id()

	ocr, err := recognizer.mc.ProcessFile(bytes.NewReader(fileBytes), fileId, mistral.ImageUrl)
	if err != nil {
		return res, wrapError(err, "mistral.ProcessFile")
	}

	ocrData, _ := json.Marshal(ocr)
	newDocumentId, savingErr := rep.CreateDocument(ctx, documentParams{
		fileId: fileId,
		chatId: chatId,
		hash:   hash,
		ocr:    ocrData,
	})

	if savingErr != nil {
		recognizer.logger.Error(wrapError(savingErr, "CreateDocument").Error())
	}

	if savingErr == nil {
		go func() {
			file, err := os.CreateTemp("", "*")

			defer func(err error) {
				if err != nil {
					recognizer.logger.Error(err.Error())
				}
			}(err)

			if err != nil {
				err = wrapError(err, "os.CreateTemp")
				return
			}
			defer func() {
				_ = file.Close()
				_ = os.Remove(file.Name())
			}()

			_, _ = io.Copy(file, bytes.NewReader(fileBytes))

			err = recognizer.s3.UploadFromLocal(ctx, file.Name(), fmt.Sprintf("%d%s", newDocumentId, path.Ext(userFile.Path())))
			if err != nil {
				err = wrapError(err, "s3.UploadFromLocal")
				return
			}

			err = rep.Commit(ctx)
			if err != nil {
				err = wrapError(err, "tx.Commit")
			}
		}()
	}

	res, _ = getOCRText(*ocr)

	return res, nil
}

func getFileCheckSum(file []byte) [16]byte {
	return md5.Sum(file)
}

func getOCRText(ocr mistral.OCRResponse) (string, bool) {
	if len(ocr.Pages) == 0 {
		return "", false
	}

	return ocr.Pages[0].Markdown, true
}

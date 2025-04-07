package mistral

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"tele/internal/config"
)

const (
	ocrEndpoint   = "https://api.mistral.ai/v1/ocr"
	filesEndpoint = "https://api.mistral.ai/v1/files"
)

const ocrModel = "mistral-ocr-latest"

type Client struct {
	cfg    config.MistralConfig
	client *http.Client
}

func New(cfg config.MistralConfig) Client {
	return Client{
		client: &http.Client{},
		cfg:    cfg,
	}
}

func (client Client) Upload(file io.Reader, fileName string) (res UploadResponse, err error) {
	const errPrefix = "client.Upload"

	var result UploadResponse

	requestParams := map[string]string{
		"purpose": "ocr",
	}

	request, err := newFileUploadRequest(filesEndpoint, file, fileName, requestParams)
	if err != nil {
		return result, fmt.Errorf("%s: make request: %w", errPrefix, err)
	}

	request.Header.Set("Authorization", "Bearer "+client.cfg.Token)

	resp, err := client.client.Do(request)
	if err != nil {
		return result, fmt.Errorf("%s: send request: %w", errPrefix, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("%s: read response: %w", errPrefix, err)
	}

	return result, nil
}

func (client Client) GetSignedURL(ctx context.Context, fileUUID string) (SignedURLResponse, error) {
	const errPrefix = "client.GetSignedURL"

	var result SignedURLResponse

	urlPath, _ := url.JoinPath(filesEndpoint, fileUUID, "url")
	uri, _ := url.Parse(urlPath)
	uri.RawQuery = "expiry=24"

	request, err := newRequest(ctx, uri.String(), http.MethodGet, nil, client.cfg.Token)
	if err != nil {
		return result, fmt.Errorf("%s: make request: %w", errPrefix, err)
	}

	result, _, err = sendAndReadResponse[SignedURLResponse](client.client, request)
	if err != nil {
		return result, fmt.Errorf("%s: read response: %w", errPrefix, err)
	}

	return result, nil
}

func (client Client) GetOCRResult(ctx context.Context, uri string, docType documentType) (OCRResponse, error) {
	const errPrefix = "client.GetOCRResult"

	var result OCRResponse

	params := Request{
		Model: ocrModel,
		Document: map[string]any{
			"type":          docType,
			string(docType): uri,
		},
	}

	request, err := newRequest(ctx, ocrEndpoint, http.MethodPost, &params, client.cfg.Token)
	if err != nil {
		return result, fmt.Errorf("%s: make request: %w", errPrefix, err)
	}

	result, _, err = sendAndReadResponse[OCRResponse](client.client, request)
	if err != nil {
		return result, fmt.Errorf("%s: %w", errPrefix, err)
	}

	return result, nil
}

func (client Client) processFile(ctx context.Context, file io.Reader, fileName string, docType documentType) (*OCRResponse, error) {
	formatError := func(err error) error {
		return fmt.Errorf("mistral.ProcessFile %s: %w", fileName, err)
	}

	r, err := client.Upload(file, fileName)
	if err != nil {
		return nil, formatError(err)
	}

	uri, err := client.GetSignedURL(ctx, r.ID)
	if err != nil {
		return nil, formatError(err)
	}

	ocr, err := client.GetOCRResult(ctx, uri.URL, docType)
	if err != nil {
		return nil, err
	}

	return &ocr, nil
}

func (client Client) GetImageOCR(ctx context.Context, file io.Reader, fileName string) (*OCRResponse, error) {
	return client.processFile(ctx, file, fileName, imageURL)
}

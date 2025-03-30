package mistral

import (
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

type documentType string

const (
	DocumentUrl documentType = "document_url"
	ImageUrl    documentType = "image_url"
)

type Request struct {
	Model    string         `json:"model"`
	Document map[string]any `json:"document"`
}

type UploadResponse struct {
	ID       string `json:"id"`
	Bytes    int64  `json:"bytes"`
	Filename string `json:"filename"`
}

type SignedURLResponse struct {
	URL string `json:"url"`
}

type OCRResponse struct {
	Pages []struct {
		Index      int    `json:"index"`
		Markdown   string `json:"markdown"`
		Images     []any  `json:"images"`
		Dimensions struct {
			Dpi    int `json:"dpi"`
			Height int `json:"height"`
			Width  int `json:"width"`
		} `json:"dimensions"`
	} `json:"pages"`
	Model     string `json:"model"`
	UsageInfo struct {
		PagesProcessed int `json:"pages_processed"`
		DocSizeBytes   int `json:"doc_size_bytes"`
	} `json:"usage_info"`
}

func (client Client) Upload(file io.ReadCloser, fileName string) (res UploadResponse, err error) {
	var result UploadResponse
	const errPrefix = "client.Upload"

	requestParams := map[string]string{
		"purpose": "ocr",
	}
	request, err := newFileUploadRequest(filesEndpoint, file, fileName, requestParams)
	if err != nil {
		return result, fmt.Errorf("%s: make request: %v", errPrefix, err)
	}

	request.Header.Set("Authorization", "Bearer "+client.cfg.Token)

	resp, err := client.client.Do(request)
	if err != nil {
		return result, fmt.Errorf("%s: send request: %v", errPrefix, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("%s: read response: %v", errPrefix, err)
	}

	return result, nil
}

func (client Client) GetSignedURL(fileUuid string) (SignedURLResponse, error) {
	const errPrefix = "client.GetSignedURL"
	var result SignedURLResponse

	urlPath, _ := url.JoinPath(filesEndpoint, fileUuid, "url")
	uri, _ := url.Parse(urlPath)
	uri.RawQuery = "expiry=24"

	request, err := newRequest[any](uri.String(), http.MethodGet, nil, client.cfg.Token)
	if err != nil {
		return result, fmt.Errorf("%s: make request: %w", errPrefix, err)
	}

	result, _, err = sendAndReadResponse[SignedURLResponse](client.client, request)
	if err != nil {
		return result, fmt.Errorf("%s: read response: %w", errPrefix, err)
	}

	return result, nil
}

func (client Client) GetOCRResult(uri string, docType documentType) (OCRResponse, error) {
	var result OCRResponse
	const errPrefix = "client.GetOCRResult"

	params := Request{
		Model: ocrModel,
		Document: map[string]any{
			"type":          docType,
			string(docType): uri,
		},
	}

	request, err := newRequest(ocrEndpoint, http.MethodPost, &params, client.cfg.Token)
	if err != nil {
		return result, fmt.Errorf("%s: make request: %v", errPrefix, err)
	}

	result, _, err = sendAndReadResponse[OCRResponse](client.client, request)
	if err != nil {
		return result, fmt.Errorf("%s: %v", errPrefix, err)
	}

	return result, nil
}

func (client Client) ProcessFile(file io.ReadCloser, fileName string, docType documentType) (*OCRResponse, error) {
	formatError := func(err error) error {
		return fmt.Errorf("mistral.ProcessFile %s: %w", fileName, err)
	}
	r, err := client.Upload(file, fileName)
	if err != nil {
		return nil, formatError(err)
	}

	uri, err := client.GetSignedURL(r.ID)
	if err != nil {
		return nil, formatError(err)
	}

	ocr, err := client.GetOCRResult(uri.URL, docType)
	if err != nil {
		return nil, err
	}

	return &ocr, nil
}

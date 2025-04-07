package mistral

type documentType string

const (
	_        documentType = "document_url"
	imageURL documentType = "image_url"
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
	Model string `json:"model"`
	//nolint:tagliatelle
	UsageInfo struct {
		PagesProcessed int `json:"pages_processed"`
		DocSizeBytes   int `json:"doc_size_bytes"`
	} `json:"usage_info"`
}

func (res *OCRResponse) Text() string {
	if len(res.Pages) == 0 {
		return ""
	}

	return res.Pages[0].Markdown
}

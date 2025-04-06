package mistral

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"io"
	"mime/multipart"
	"net/http"
)

var retryStatusCodes = map[int]struct{}{
	http.StatusTooManyRequests:     {},
	http.StatusInternalServerError: {},
	http.StatusBadGateway:          {},
	http.StatusServiceUnavailable:  {},
	http.StatusGatewayTimeout:      {},
}

var recoverableHttpStatusError = errors.New("recoverable http status")

func newRequest(
	uri string,
	method string,
	params any,
	apiToken string,
) (req *http.Request, err error) {
	body := &bytes.Buffer{}

	if params != nil {
		_ = json.NewEncoder(body).Encode(params)
	}

	req, err = http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func newFileUploadRequest(
	uri string,
	file io.Reader,
	fileName string,
	params map[string]string,
) (*http.Request, error) {
	const fileField = "file"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	part, err := writer.CreateFormFile(fileField, fileName)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func sendRequestWithRetry(httpClient *http.Client, request *http.Request) (http.Response, error) {
	var res http.Response
	err := retry.Do(func() error {
		response, err := httpClient.Do(request)
		if err != nil {
			return err
		}

		res = *response

		if _, ok := retryStatusCodes[res.StatusCode]; ok {
			return recoverableHttpStatusError
		} else if res.StatusCode >= 400 {
			return errors.New(fmt.Sprintf("response status: %s", res.Status))
		}

		return nil
	}, retry.RetryIf(func(err error) bool {
		return errors.Is(err, recoverableHttpStatusError)
	}))

	return res, err
}

func sendAndReadResponse[T any](httpClient *http.Client, request *http.Request) (T, *int, error) {
	var result T

	resp, err := sendRequestWithRetry(httpClient, request)
	if err != nil {
		return result, nil, fmt.Errorf("send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, &resp.StatusCode, fmt.Errorf("read response: %v", err)
	}

	return result, &resp.StatusCode, nil
}

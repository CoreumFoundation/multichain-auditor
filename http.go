package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

func DoJSON(ctx context.Context, method, url string, reqBody, respBody interface{}) error {
	var reqBodyReader io.Reader
	if reqBody != nil {
		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return errors.Errorf("can't marshal request body, err: %v", err)
		}
		reqBodyReader = bytes.NewReader(reqBodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBodyReader)
	if err != nil {
		return errors.Errorf("can't build the request, err: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Errorf("can't perform the request, err: %v", err)
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("can't read the response body, err: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.Errorf("can't perform request, code: %d, body: %s", resp.StatusCode, string(bodyData))
	}

	err = json.Unmarshal(bodyData, respBody)
	if err != nil {
		return errors.Errorf("can't unmarshal the response body, body: %s, err: %v", string(bodyData), err)
	}

	return nil
}

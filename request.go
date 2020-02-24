package OpenTracing

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

// Request - for http requests. Expect method, url and body. Will inject bearer token and tracing id to request.
func Request(ctx context.Context, method string, url string, b []byte, auth bool) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	if err := InjectToReq(ctx, req); err != nil {
		return nil, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		log.Printf("[DEBUG] %s %s: %d (%s) %s", method, url, resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
		return nil, errors.New(string(body))
	}

	return body, nil
}

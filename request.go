package OpenTracing

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RequestOpts struct {
	Client       http.Client
	Method       string
	URL          string
	Data         []byte
	InjectBearer bool
}

// Request - for http requests. Expect method, url and body. Will inject bearer token and tracing id to request.
func Request(ctx context.Context, params RequestOpts) ([]byte, int, error) {
	req, err := http.NewRequest(params.Method, params.URL, bytes.NewBuffer(params.Data))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Add("Content-Type", "application/json")
	if params.InjectBearer {
		if err := SetAuthorization(ctx, req); err != nil {
			return nil, 0, err
		}
	}
	if err := InjectToReq(ctx, req); err != nil {
		return nil, 0, err
	}

	resp, err := params.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, errors.New(string(body))
	}

	return body, resp.StatusCode, nil
}

// SetAuthorization - take token from context value and set as Bearer token in Authorization header.
func SetAuthorization(ctx context.Context, req *http.Request) error {
	jwtToken := ctx.Value("token")
	if jwtToken == nil || jwtToken == "" {
		return errors.New("no token in context")
	}

	if req == nil {
		return errors.New("empty request")
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	return nil
}

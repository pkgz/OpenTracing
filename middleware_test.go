package OpenTracing

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenTracing(t *testing.T) {
	tracer, closer, err := NewTracer("test", "")
	require.NoError(t, err)
	require.NotNil(t, tracer)
	require.NotNil(t, closer)

	type response struct {
		SpanID string
	}
	ts := httptest.NewServer(OpenTracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		b, _ := json.Marshal(response{
			SpanID: fmt.Sprintf("%v", opentracing.SpanFromContext(r.Context())),
		})
		_, _ = w.Write(b)
	})))
	defer ts.Close()

	t.Run("no span", func(t *testing.T) {
		resp, err := http.Get(ts.URL)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		require.NotEmpty(t, body)
		require.NoError(t, err)

		var response response
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)

		require.NotEmpty(t, response.SpanID)
	})

	t.Run("span continuation", func(t *testing.T) {
		ctx := context.Background()
		req, err := http.NewRequest("GET", ts.URL, nil)
		require.NoError(t, err)

		span, ctx := opentracing.StartSpanFromContext(ctx, "test")
		require.NoError(t, InjectToReq(ctx, req))

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		require.NotEmpty(t, body)
		require.NoError(t, err)

		var response response
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)

		require.NotEmpty(t, response.SpanID)

		localSpanArr := strings.Split(fmt.Sprintf("%v", span), ":")
		remoteSpanArr := strings.Split(response.SpanID, ":")

		require.Equal(t, localSpanArr[0], remoteSpanArr[0])
		require.Equal(t, localSpanArr[1], remoteSpanArr[2])
	})
}

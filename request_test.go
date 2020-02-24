package OpenTracing

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(nil)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer ts.Close()

	t.Run("invalid method", func(t *testing.T) {
		resp, err := Request(context.Background(), " ", "", []byte{}, false)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("no token", func(t *testing.T) {
		resp, err := Request(context.Background(), "", "", []byte{}, true)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("no span", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "test")
		resp, err := Request(ctx, "", "", []byte{}, true)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("wrong url", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "test")
		span, ctx := opentracing.StartSpanFromContext(ctx, "test")
		require.NotNil(t, span)

		resp, err := Request(ctx, "", "", []byte{}, true)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("non 1xx, 2xx, 3xx response", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "test")
		span, ctx := opentracing.StartSpanFromContext(ctx, "test")
		require.NotNil(t, span)

		resp, err := Request(ctx, "GET", ts.URL, []byte{}, true)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("200", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "test")
		span, ctx := opentracing.StartSpanFromContext(ctx, "test")
		require.NotNil(t, span)

		resp, err := Request(ctx, "POST", ts.URL, []byte{}, true)
		require.NoError(t, err)
		require.Equal(t, []byte("OK"), resp)
	})
}

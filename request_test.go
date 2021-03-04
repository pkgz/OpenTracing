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
		resp, code, err := Request(context.Background(), RequestOpts{})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Equal(t, 0, code)
	})

	t.Run("no token", func(t *testing.T) {
		resp, code, err := Request(context.Background(), RequestOpts{
			InjectBearer: true,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Equal(t, 0, code)
	})

	t.Run("no span", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "test")
		resp, code, err := Request(ctx, RequestOpts{
			InjectBearer: true,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Equal(t, 0, code)
	})

	t.Run("wrong url", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "test")
		span, ctx := opentracing.StartSpanFromContext(ctx, "test")
		require.NotNil(t, span)

		resp, code, err := Request(ctx, RequestOpts{
			InjectBearer: true,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Equal(t, 0, code)
	})

	t.Run("non 1xx, 2xx, 3xx response", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "test")
		span, ctx := opentracing.StartSpanFromContext(ctx, "test")
		require.NotNil(t, span)

		resp, code, err := Request(ctx, RequestOpts{
			Method:       "GET",
			URL:          ts.URL,
			InjectBearer: true,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Equal(t, http.StatusBadRequest, code)
	})

	t.Run("200", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "test")
		span, ctx := opentracing.StartSpanFromContext(ctx, "test")
		require.NotNil(t, span)

		resp, code, err := Request(ctx, RequestOpts{
			Method:       "POST",
			URL:          ts.URL,
			InjectBearer: true,
		})
		require.NoError(t, err)
		require.Equal(t, []byte("OK"), resp)
		require.Equal(t, http.StatusOK, code)
	})
}

func TestSetAuthorization(t *testing.T) {
	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		require.Error(t, SetAuthorization(ctx, nil))
	})
	t.Run("empty request", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "token", "test")
		require.Error(t, SetAuthorization(ctx, nil))
	})
	t.Run("ok", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "token", "test")
		req := httptest.NewRequest("GET", "http://test.com", nil)
		require.NoError(t, SetAuthorization(ctx, req))
	})
}

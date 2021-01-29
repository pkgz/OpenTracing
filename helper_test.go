package OpenTracing

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/require"
	"github.com/uber/jaeger-client-go"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTracer(t *testing.T) {
	ctx := context.Background()

	tracer, err := NewTracer(ctx, "test", "")
	require.NoError(t, err)
	require.NotNil(t, tracer)

	tracer, err = NewTracer(ctx, "test", "wtf")
	require.Error(t, err)
}

func TestInjectToReq(t *testing.T) {
	tracer, err := NewTracer(context.Background(), "test", "")
	require.NoError(t, err)
	require.NotNil(t, tracer)

	handler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("Uber-Trace-Id")))
	}

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		req := httptest.NewRequest("GET", "http://test.com", nil)
		err := InjectToReq(ctx, req)
		require.Error(t, err)
	})

	t.Run("create span", func(t *testing.T) {
		ctx := context.Background()
		span, ctx := opentracing.StartSpanFromContext(ctx, "test")
		require.NotNil(t, span)

		req := httptest.NewRequest("GET", "http://test.com", nil)
		err := InjectToReq(ctx, req)
		require.NoError(t, err)

		span.Finish()
	})

	t.Run("request with span", func(t *testing.T) {
		ctx := context.Background()
		span, ctx := opentracing.StartSpanFromContext(ctx, "test")
		require.NotNil(t, span)

		req := httptest.NewRequest("GET", "http://test.com", nil)
		err := InjectToReq(ctx, req)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		sc, ok := span.Context().(jaeger.SpanContext)
		require.True(t, ok)
		require.NotNil(t, sc)

		require.Equal(t, sc.String(), string(body))

		span.Finish()
	})
}

func TestError(t *testing.T) {
	ctx := context.Background()
	span, ctx := opentracing.StartSpanFromContext(ctx, "test")
	Error(span)
}
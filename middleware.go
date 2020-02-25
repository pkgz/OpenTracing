package OpenTracing

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/http"
)

// Middleware - create a new openTracing span and put it into request context
func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var span opentracing.Span
		ctx := r.Context()

		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			span, ctx = opentracing.StartSpanFromContext(r.Context(), fmt.Sprintf("%s: %s", r.Method, r.URL.Path))
		} else {
			span, ctx = opentracing.StartSpanFromContext(r.Context(), fmt.Sprintf("%s: %s", r.Method, r.URL.Path), opentracing.ChildOf(wireContext))
		}
		defer span.Finish()

		if err := span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier); err != nil {
			span.SetTag(string(ext.Error), true)
		} else {
			r = r.WithContext(opentracing.ContextWithSpan(ctx, span))
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

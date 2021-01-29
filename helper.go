package OpenTracing

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"log"
	"net/http"
)

// NewTracer - initializing opentracing tracer and set up global tracer.
func NewTracer(ctx context.Context, name string, host string) (opentracing.Tracer, error) {
	udpTransport, err := jaeger.NewUDPTransport(fmt.Sprintf("%s:6831", host), 0)
	if err != nil {
		return nil, errors.Wrap(err, "create transport error")
	}

	cfg := jaegercfg.Configuration{
		ServiceName: name,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	reporter := jaeger.NewRemoteReporter(udpTransport)

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Reporter(reporter),
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)

	go func() {
		<-ctx.Done()
		if err = closer.Close(); err != nil {
			log.Panic(err)
		}
	}()

	opentracing.SetGlobalTracer(tracer)

	return tracer, nil
}

// InjectToReq - inject opentracing span to *http.Requst
func InjectToReq(ctx context.Context, req *http.Request) error {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return errors.New("no span in context")
	}

	return opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
}

// Error - finish a span and set error status
func Error(span opentracing.Span) {
	span.SetTag(string(ext.Error), true)
	span.Finish()
}
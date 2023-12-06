package telemetry

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger" //nolint:staticcheck // This is deprecated and will be removed in the next release.
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Service interface {
	Start(ctx context.Context, spanName string) (context.Context, Span)
	TracerProviderOption() otelhttp.Option
}

func newTraceProvider(_ context.Context, telemetryAddr string, serviceName string) (trace.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(telemetryAddr)))

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new tracer")
	}
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)

	if err != nil {
		return nil, err
	}
	return sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(r),
		),
		nil
}

func NewService(ctx context.Context, telemetryAddr string, serviceName string) (Service, error) {
	tracerProvider, err := newTraceProvider(ctx, telemetryAddr, serviceName)
	if err != nil {
		return nil, err
	}

	tracer := tracerProvider.Tracer(serviceName)
	if err != nil {
		return nil, err
	}

	return jaegerTracer{tracer: tracer, tracerProvider: tracerProvider}, nil
}

type jaegerTracer struct {
	tracer         trace.Tracer
	tracerProvider trace.TracerProvider
}

func (j jaegerTracer) Start(ctx context.Context, spanName string) (context.Context, Span) {
	ctx, span := j.tracer.Start(ctx, spanName)
	return ctx, jaegerSpan{span: span}
}

func (j jaegerTracer) TracerProviderOption() otelhttp.Option {
	return otelhttp.WithTracerProvider(j.tracerProvider)
}

type Span interface {
	AddEvent(EventName)
	End()
	SetTag(LabelKey, LabelValue)
	SetError(error)
}

type jaegerSpan struct {
	span trace.Span
}

type LabelKey = attribute.Key
type LabelValue = string
type EventName = string

func (j jaegerSpan) AddEvent(eventName EventName) {
	j.span.AddEvent(eventName)
}

func (j jaegerSpan) End() {
	j.span.End()
}

func (j jaegerSpan) SetTag(key LabelKey, value LabelValue) {
	j.span.SetAttributes(key.String(value))
}

func (j jaegerSpan) SetError(err error) {
	j.span.RecordError(err)
}

func GetSpanFromContext(ctx context.Context) Span {
	return jaegerSpan{trace.SpanFromContext(ctx)}
}

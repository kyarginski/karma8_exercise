package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger" //nolint:staticcheck // This is deprecated and will be removed in the next release.
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	httptrace "go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
)

var (
	UseTracing   bool
	ServiceTitle string
)

// InitJaegerTracer Инициализация OpenTelemetry для Jaeger trace.
func InitJaegerTracer(url string, serviceName string, environment string, useTracing bool) (*tracesdk.TracerProvider, error) {
	ServiceTitle = serviceName
	UseTracing = useTracing
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			attribute.String("environment", environment),
		)),
	)
	return tp, nil
}

// Получить строковое значение TraceId (A valid trace identifier is a 16-byte array with at least one non-zero byte).
func getNewTraceID() string {
	res := getSizedRandomBytes(16)

	return hex.EncodeToString(res)
}

// Получить строковое значение SpanId (A valid span identifier is an 8-byte array with at least one non-zero byte).
func getNewSpanID() string {
	res := getSizedRandomBytes(8)

	return hex.EncodeToString(res)
}

// Получить случайную последовательность байт размером size.
func getSizedRandomBytes(size int) []byte {
	token := make([]byte, size)
	_, _ = rand.Read(token)

	return token
}

// CreateMasterSpan Создать основной span.
func CreateMasterSpan(r *http.Request) (context.Context, trace.Span) {
	parentTracer := otel.Tracer(ServiceTitle)
	attrs, _, spanCtx := httptrace.Extract(r.Context(), r)

	// traceID SpanId
	traceID := r.Header.Get("TraceId")
	spanID := r.Header.Get("SpanId")
	if traceID != "" && spanID != "" {
		attrs = append(attrs, attribute.String("TraceId", traceID))
		attrs = append(attrs, attribute.String("SpanId", spanID))
	} else {
		traceID = getNewTraceID()
		spanID = getNewSpanID()
	}
	mySpanID, _ := trace.SpanIDFromHex(spanID)
	myTraceID, _ := trace.TraceIDFromHex(traceID)
	spanCtx = spanCtx.WithSpanID(mySpanID).WithTraceID(myTraceID)

	var span trace.Span
	traceContext, span := parentTracer.Start(
		trace.ContextWithRemoteSpanContext(r.Context(), spanCtx),
		r.RequestURI,
		trace.WithAttributes(attrs...),
	)

	return traceContext, span
}

func CreateSubSpan(ctx context.Context, r *http.Request, name string) trace.Span {
	tr := otel.Tracer(ServiceTitle)
	_, span := tr.Start(ctx, name)

	// Obtain the span context from the span
	spanCtx := span.SpanContext()
	traceID := spanCtx.TraceID().String()
	spanID := spanCtx.SpanID().String()

	// если не traceID и spanID не пришли от клиента (нулевые значения), то задаём их сами
	if traceID == (trace.TraceID{}).String() || spanID == (trace.SpanID{}).String() {
		traceID = getNewTraceID()
		spanID = getNewSpanID()
	}
	r.Header.Set("TraceId", traceID)
	r.Header.Set("SpanId", spanID)

	return span
}

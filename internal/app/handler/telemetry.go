package handler

import (
	"context"
	"net/http"

	trccontext "karma8/internal/lib/context"
	"karma8/internal/lib/monitoring/telemetry"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
)

func TelemetryFromConfig(ctx context.Context, telemetryAddr string, serviceName string) (func(http.Handler) http.Handler, error) {
	tracer, err := telemetry.NewService(ctx, telemetryAddr, serviceName)
	if err != nil {
		return noopMiddleware, err
	}
	return Telemetry(tracer, serviceName), nil
}

func Telemetry(telemetr telemetry.Service, name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		setTelemetryLabelsHandler := func(rw http.ResponseWriter, r *http.Request) {
			r = r.WithContext(trccontext.WithTelemetry(r.Context(), telemetr))

			span := trccontext.CurrentSpanFromContext(r.Context())

			requestID, ok := trccontext.RequestIDFromContext(r.Context())
			if !ok {
				requestID = "UNKNOWN"
			}
			span.SetTag("requestID", requestID)

			span.SetTag("path", r.URL.Path)

			next.ServeHTTP(rw, r)
		}

		healthEndpointFilter := func(r *http.Request) bool {
			if r.URL == nil {
				return true
			}
			return r.URL.Path != "/health"
		}

		return otelhttp.NewHandler(
			http.HandlerFunc(setTelemetryLabelsHandler),
			name+"_handler",
			telemetr.TracerProviderOption(),
			otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})),
			otelhttp.WithFilter(healthEndpointFilter),
		)
	}
}

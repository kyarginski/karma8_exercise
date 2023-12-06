package context

import (
	"context"

	"karma8/internal/lib/monitoring/telemetry"
)

type (
	ctxKeyJobID           struct{}
	ctxKeyJobName         struct{}
	ctxKeyRequestID       struct{}
	ctxKeyAppInstanceName struct{}
	ctxKeyTelemetry       struct{}

	Data struct {
		JobID              string
		JobIDSet           bool
		JobName            string
		JobNameSet         bool
		RequestID          string
		RequestIDSet       bool
		ProviderName       string
		ProviderNameSet    bool
		CasinoName         string
		CasinoNameSet      bool
		AppInstanceName    string
		AppInstanceNameSet bool
	}
)

func GetContextData(ctx context.Context) Data {
	data := Data{}

	data.JobID, data.JobIDSet = JobIDFromContext(ctx)
	data.JobName, data.JobNameSet = JobNameFromContext(ctx)
	data.RequestID, data.RequestIDSet = RequestIDFromContext(ctx)
	data.AppInstanceName, data.AppInstanceNameSet = AppInstanceNameFromContext(ctx)

	return data
}

func WithData(ctxParent context.Context, data Data) (ctxChild context.Context) {
	ctxChild = ctxParent
	if data.JobIDSet {
		ctxChild = WithJobID(ctxChild, data.JobID)
	}
	if data.JobNameSet {
		ctxChild = WithJobName(ctxChild, data.JobName)
	}
	if data.RequestIDSet {
		ctxChild = WithRequestID(ctxChild, data.RequestID)
	}
	if data.AppInstanceNameSet {
		ctxChild = WithAppInstanceName(ctxChild, data.AppInstanceName)
	}
	return ctxChild
}

func CopyContextData(ctxParent, ctxWithData context.Context) (ctxChild context.Context) {
	return WithData(ctxParent, GetContextData(ctxWithData))
}

// Job ID.

func WithJobID(ctx context.Context, jobID string) context.Context {
	return context.WithValue(ctx, ctxKeyJobID{}, jobID)
}

func JobIDFromContext(ctx context.Context) (jobID string, ok bool) {
	if ctx == nil {
		return
	}
	jobID, ok = ctx.Value(ctxKeyJobID{}).(string)
	return
}

// Job name.

func WithJobName(ctx context.Context, jobName string) context.Context {
	return context.WithValue(ctx, ctxKeyJobName{}, jobName)
}

func JobNameFromContext(ctx context.Context) (jobName string, ok bool) {
	if ctx == nil {
		return
	}
	jobName, ok = ctx.Value(ctxKeyJobName{}).(string)
	return
}

// Request ID.

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID{}, requestID)
}

func RequestIDFromContext(ctx context.Context) (requestID string, ok bool) {
	if ctx == nil {
		return
	}
	requestID, ok = ctx.Value(ctxKeyRequestID{}).(string)
	return
}

// App instance name.

func WithAppInstanceName(ctx context.Context, appInstanceName string) context.Context {
	return context.WithValue(ctx, ctxKeyAppInstanceName{}, appInstanceName)
}

func AppInstanceNameFromContext(ctx context.Context) (appInstanceName string, ok bool) {
	if ctx == nil {
		return
	}
	appInstanceName, ok = ctx.Value(ctxKeyAppInstanceName{}).(string)
	return
}

func WithTelemetry(ctx context.Context, service telemetry.Service) context.Context {
	if service == nil || ctx == nil {
		return ctx
	}

	return context.WithValue(ctx, ctxKeyTelemetry{}, service)
}

func telemetryFromContext(ctx context.Context) (telemetry.Service, bool) {
	if ctx == nil {
		return nil, false
	}
	result, ok := ctx.Value(ctxKeyTelemetry{}).(telemetry.Service)
	return result, ok
}

type noopSpan struct{}

func (noopSpan) AddEvent(_ telemetry.EventName)                  {}
func (noopSpan) End()                                            {}
func (noopSpan) SetTag(telemetry.LabelKey, telemetry.LabelValue) {}
func (noopSpan) SetError(error)                                  {}

func WithTelemetrySpan(parent context.Context, spanName string) (ctx context.Context, span telemetry.Span) {
	telemetryService, ok := telemetryFromContext(parent)
	if !ok {
		return parent, noopSpan{}
	}
	return telemetryService.Start(parent, spanName)
}

func CurrentSpanFromContext(ctx context.Context) telemetry.Span {
	return telemetry.GetSpanFromContext(ctx)
}

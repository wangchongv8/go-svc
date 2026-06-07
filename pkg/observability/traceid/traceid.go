// Package traceid provides helpers to extract trace id and span id
// from an OpenTelemetry context.
package traceid

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
)

// FromContext returns the current trace id and span id from context.
// Returns empty strings if no valid span is present.
func FromContext(ctx context.Context) (traceID, spanID string) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return "", ""
	}

	tid := span.SpanContext().TraceID()
	sid := span.SpanContext().SpanID()
	return tid.String(), sid.String()
}

// LogFields returns logx.Field slices for trace id and span id.
// Use with logx.Infow/Errorw for structured trace correlation.
func LogFields(ctx context.Context) []logx.LogField {
	tid, sid := FromContext(ctx)
	fields := make([]logx.LogField, 0, 2)
	if tid != "" {
		fields = append(fields, logx.Field("trace_id", tid))
	}
	if sid != "" {
		fields = append(fields, logx.Field("span_id", sid))
	}
	return fields
}

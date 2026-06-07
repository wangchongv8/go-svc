// Package traceid provides helpers to extract trace id and span id
// from an OpenTelemetry context.
package traceid

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

const (
	traceIDMetadataKey = "x-trace-id"
	spanIDMetadataKey  = "x-span-id"
)

// FromContext returns the current trace id and span id from context.
// Returns empty strings if no valid span is present.
func FromContext(ctx context.Context) (traceID, spanID string) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		tid := firstMetadataValue(md, traceIDMetadataKey)
		if tid != "" {
			return tid, firstMetadataValue(md, spanIDMetadataKey)
		}
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return "", ""
	}

	tid := span.SpanContext().TraceID()
	sid := span.SpanContext().SpanID()
	return tid.String(), sid.String()
}

// WithOutgoingMetadata propagates the current trace id and span id through
// gRPC metadata so downstream services can write logs with the caller trace id.
func WithOutgoingMetadata(ctx context.Context) context.Context {
	tid, sid := FromContext(ctx)
	if tid == "" {
		return ctx
	}
	if sid == "" {
		return metadata.AppendToOutgoingContext(ctx, traceIDMetadataKey, tid)
	}
	return metadata.AppendToOutgoingContext(ctx, traceIDMetadataKey, tid, spanIDMetadataKey, sid)
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

func firstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

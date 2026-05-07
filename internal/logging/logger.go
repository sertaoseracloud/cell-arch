package logging

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// NewLogger creates a JSON-structured logger that includes trace context.
func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// WithTrace returns slog.Attr slices containing trace_id and span_id from the context.
func WithTrace(ctx context.Context) []slog.Attr {
	span := trace.SpanFromContext(ctx).SpanContext()
	if !span.IsValid() {
		return nil
	}
	return []slog.Attr{
		slog.String("trace_id", span.TraceID().String()),
		slog.String("span_id", span.SpanID().String()),
	}
}

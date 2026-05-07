// Package contextutil provides helpers for working with context.Context
// across service boundaries.
package contextutil

import "context"

type contextKey string

const (
	requestIDKey  contextKey = "request_id"
	correlationKey contextKey = "correlation_id"
)

// WithRequestID attaches a request ID to the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFrom extracts the request ID from the context.
// Returns an empty string if not set.
func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// WithCorrelationID attaches a correlation ID to the context for
// cross-service trace propagation.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationKey, id)
}

// CorrelationIDFrom extracts the correlation ID from the context.
// Returns an empty string if not set.
func CorrelationIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(correlationKey).(string); ok {
		return v
	}
	return ""
}

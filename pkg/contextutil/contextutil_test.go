package contextutil_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/cell-arch/pkg/contextutil"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name    string
		setID   string
		want    string
		skipSet bool
	}{
		{
			name:  "stores and retrieves request ID",
			setID: "req-abc-123",
			want:  "req-abc-123",
		},
		{
			name:    "returns empty string when not set",
			skipSet: true,
			want:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if !tc.skipSet {
				ctx = contextutil.WithRequestID(ctx, tc.setID)
			}
			got := contextutil.RequestIDFrom(ctx)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCorrelationID(t *testing.T) {
	tests := []struct {
		name    string
		setID   string
		want    string
		skipSet bool
	}{
		{
			name:  "stores and retrieves correlation ID",
			setID: "corr-xyz-456",
			want:  "corr-xyz-456",
		},
		{
			name:    "returns empty string when not set",
			skipSet: true,
			want:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if !tc.skipSet {
				ctx = contextutil.WithCorrelationID(ctx, tc.setID)
			}
			got := contextutil.CorrelationIDFrom(ctx)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBothIDs_Independent(t *testing.T) {
	ctx := context.Background()
	ctx = contextutil.WithRequestID(ctx, "req-1")
	ctx = contextutil.WithCorrelationID(ctx, "corr-1")

	assert.Equal(t, "req-1", contextutil.RequestIDFrom(ctx))
	assert.Equal(t, "corr-1", contextutil.CorrelationIDFrom(ctx))
}

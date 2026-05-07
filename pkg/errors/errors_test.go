package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pkgerrors "github.com/yourorg/cell-arch/pkg/errors"
)

var errSentinel = stderrors.New("sentinel error")

// customErr is a typed error for testing errors.As behaviour.
type customErr struct{ msg string }

func (e *customErr) Error() string { return e.msg }

func TestWrap(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		msg         string
		wantNil     bool
		errContains string
		wantIs      error
	}{
		{
			name:        "wraps error with message",
			err:         errSentinel,
			msg:         "context message",
			wantNil:     false,
			errContains: "context message",
			wantIs:      errSentinel,
		},
		{
			name:    "returns nil when err is nil",
			err:     nil,
			msg:     "message",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := pkgerrors.Wrap(tc.err, tc.msg)
			if tc.wantNil {
				assert.Nil(t, got)
				return
			}
			require.Error(t, got)
			assert.Contains(t, got.Error(), tc.errContains)
			if tc.wantIs != nil {
				assert.True(t, pkgerrors.Is(got, tc.wantIs), "errors.Is should find wrapped sentinel")
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		format      string
		args        []any
		wantNil     bool
		errContains string
	}{
		{
			name:        "wraps with formatted message",
			err:         errSentinel,
			format:      "operation %s failed for id %d",
			args:        []any{"create", 42},
			wantNil:     false,
			errContains: "create",
		},
		{
			name:    "returns nil when err is nil",
			err:     nil,
			format:  "msg",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := pkgerrors.Wrapf(tc.err, tc.format, tc.args...)
			if tc.wantNil {
				assert.Nil(t, got)
				return
			}
			require.Error(t, got)
			assert.Contains(t, got.Error(), tc.errContains)
			assert.True(t, pkgerrors.Is(got, errSentinel))
		})
	}
}

func TestNew(t *testing.T) {
	err := pkgerrors.New("brand new error")
	require.Error(t, err)
	assert.Equal(t, "brand new error", err.Error())
}

func TestAs(t *testing.T) {
	orig := &customErr{msg: "custom"}
	wrapped := pkgerrors.Wrap(orig, "outer")

	var target *customErr
	assert.True(t, pkgerrors.As(wrapped, &target))
	assert.Equal(t, "custom", target.msg)
}

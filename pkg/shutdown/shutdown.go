// Package shutdown provides helpers for implementing graceful shutdown in
// Go services. It follows D-04: OS signal handling with context.WithCancel.
package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
)

// Graceful creates a context that is cancelled when the process receives
// SIGINT or SIGTERM. The returned stop function must be called with defer
// to release signal resources.
//
// Usage:
//
//	ctx, stop := shutdown.Graceful(parentCtx, logger)
//	defer stop()
//	// ... start servers ...
//	<-ctx.Done()  // blocks until signal received
func Graceful(parent context.Context, logger zerolog.Logger) (context.Context, context.CancelFunc) {
	// sigCh is registered BEFORE NotifyContext so we capture the signal that
	// triggers cancellation. NotifyContext consumes the signal internally and
	// cancels ctx; sigCh receives the same signal concurrently.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)

	// Spawn a goroutine that logs when the signal is received so that
	// operational logs show the shutdown reason.
	go func() {
		select {
		case sig := <-sigCh:
			logger.Info().
				Str("signal", sig.String()).
				Msg("OS signal received — initiating graceful shutdown")
		case <-ctx.Done():
			// Context cancelled for a non-signal reason (e.g. stop() called directly).
		}
		signal.Stop(sigCh)
	}()

	return ctx, stop
}

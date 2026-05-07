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
	ctx, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)

	// Spawn a goroutine that logs when the signal is received so that
	// operational logs show the shutdown reason.
	go func() {
		<-ctx.Done()
		logger.Info().
			Str("signal", signalName()).
			Msg("OS signal received — initiating graceful shutdown")
	}()

	return ctx, stop
}

// signalName returns a human-readable name for the signal that triggered
// shutdown. It reads the process' signal state via os.Interrupt as a fallback.
func signalName() string {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	select {
	case sig := <-ch:
		return sig.String()
	default:
		return "unknown"
	}
}

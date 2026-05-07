// Package main is the entry point for the cell-arch main application.
// It wires all dependencies via the container (D-02), sets up structured
// logging (D-08), and handles graceful shutdown on SIGINT/SIGTERM (D-04).
//
// IMPORTANT: This package must NOT import any AWS or Azure SDK packages.
// All cloud interaction is delegated to the sidecar via gRPC on localhost.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 1. Initialise the structured logger first so all subsequent steps can log.
	logger := initLogger()

	// 2. Root context cancelled on OS signals (D-04: graceful shutdown).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Fatal().Err(err).Msg("application error")
		os.Exit(1)
	}
	logger.Info().Msg("application stopped gracefully")
}

// run contains the main application logic. Accepting ctx and logger as arguments
// makes it straightforward to test without capturing os.Exit.
func run(ctx context.Context, logger zerolog.Logger) error {
	// 3. Wire all dependencies via the application container (D-02).
	c, err := newContainer(ctx, logger)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	c.logger.Info().
		Str("sidecar_addr", c.cfg.SidecarAddr).
		Str("env", c.cfg.Env).
		Int("grpc_port", c.cfg.GRPCPort).
		Msg("application started — awaiting shutdown signal")

	// Phase 2 will start the HTTP/gRPC server using c.taskService here.
	_ = c.taskService

	// 4. Block until the OS sends SIGINT or SIGTERM.
	<-ctx.Done()
	c.logger.Info().Msg("shutdown signal received, stopping…")

	return nil
}

// initLogger creates a zerolog.Logger appropriate for the runtime environment.
// In development, output is human-readable via ConsoleWriter.
// In production (APP_ENV=production), output is raw JSON.
func initLogger() zerolog.Logger {
	if os.Getenv("APP_ENV") == "production" {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		return zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
	return log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false}).
		With().Timestamp().Logger()
}

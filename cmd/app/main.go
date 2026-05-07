// Package main is the entry point for the cell-arch main application.
// It wires all dependencies via manual constructors (D-02), sets up structured
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
	"github.com/yourorg/cell-arch/internal/config"
	"github.com/yourorg/cell-arch/internal/task/infrastructure/repo"
	"github.com/yourorg/cell-arch/internal/task/usecase"
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
	// 3. Load typed configuration from environment variables (D-02, D-03).
	cfg, err := config.LoadAppConfig(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	logger.Info().
		Str("sidecar_addr", cfg.SidecarAddr).
		Str("env", cfg.Env).
		Int("grpc_port", cfg.GRPCPort).
		Msg("configuration loaded")

	// 4. Wire the infrastructure adapter (sidecar repo) — no cloud SDKs here.
	taskRepo := repo.NewSidecarRepo(cfg.SidecarAddr, "aws", logger)

	// 5. Wire the use-case service with the repository.
	taskService := usecase.NewService(taskRepo, logger)

	// Suppress unused-variable warning until HTTP/gRPC server is wired in Phase 2.
	_ = taskService

	logger.Info().Msg("application started — awaiting shutdown signal")

	// 6. Block until the OS sends SIGINT or SIGTERM.
	<-ctx.Done()
	logger.Info().Msg("shutdown signal received, stopping…")

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

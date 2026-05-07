// Package main is the entry point for the cell-arch sidecar proxy.
// The sidecar holds all cloud SDK credentials and exposes a unified
// gRPC API on localhost:50051 for the main app to consume.
//
// Phase 2 will wire the full gRPC server with protobuf-generated handlers.
// This Phase 1 skeleton establishes the graceful-shutdown pattern (D-04),
// structured logging (D-08), and manual constructor wiring (D-02).
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yourorg/cell-arch/internal/config"
	sidecaraws "github.com/yourorg/cell-arch/internal/sidecar/aws"
	sidecarazure "github.com/yourorg/cell-arch/internal/sidecar/azure"
	"google.golang.org/grpc"
)

func main() {
	// 1. Initialise the structured logger.
	logger := initLogger()

	// 2. Root context cancelled on OS signals (D-04).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Fatal().Err(err).Msg("sidecar error")
		os.Exit(1)
	}
	logger.Info().Msg("sidecar stopped gracefully")
}

// run contains the main sidecar logic.
func run(ctx context.Context, logger zerolog.Logger) error {
	// 3. Load typed configuration.
	cfg, err := config.LoadSidecarConfig(ctx)
	if err != nil {
		return fmt.Errorf("load sidecar config: %w", err)
	}

	// 4. Initialise AWS client (IRSA — no static credentials).
	awsClient, err := sidecaraws.NewDynamoDBClient(
		ctx,
		cfg.AWSRegion,
		cfg.DynamoDBTable,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create AWS DynamoDB client: %w", err)
	}
	_ = awsClient // Phase 2 will register gRPC handlers

	// 5. Initialise Azure client (Workload Identity — no static credentials).
	azureClient, err := sidecarazure.NewCosmosDBClient(
		ctx,
		cfg.AzureCosmosEndpoint,
		cfg.CosmosDatabase,
		cfg.CosmosContainer,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create Azure CosmosDB client: %w", err)
	}
	_ = azureClient // Phase 2 will register gRPC handlers

	// 6. Start gRPC listener.
	addr := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	// Phase 1: plain gRPC server (no TLS until cert-manager is configured in Phase 5).
	// Phase 2 will add mTLS via grpc.Creds.
	grpcServer := grpc.NewServer()

	// Phase 2 TODO: register TaskServiceServer handler here.

	logger.Info().Str("addr", addr).Msg("sidecar starting")

	// 7. Serve in background; cancel on context cancellation.
	errCh := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc serve: %w", err)
		}
		close(errCh)
	}()

	// 8. Block until OS signal or serve error.
	select {
	case <-ctx.Done():
		logger.Info().Msg("shutdown signal received, draining connections…")
		grpcServer.GracefulStop()
		return <-errCh
	case err := <-errCh:
		return err
	}
}

// initLogger creates a zerolog.Logger appropriate for the runtime environment.
func initLogger() zerolog.Logger {
	if os.Getenv("APP_ENV") == "production" {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		return zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
	return log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false}).
		With().Timestamp().Logger()
}

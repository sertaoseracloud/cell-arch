// Package main is the entry point for the cell-arch sidecar proxy.
// The sidecar holds all cloud SDK credentials and exposes a unified
// gRPC API on localhost:50051 for the main app to consume.
//
// Design decisions honoured:
//   - D-09: gRPC only, no HTTP/REST.
//   - D-11: grpc-go server framework.
//   - D-15: mTLS between app and sidecar (cert/key loaded from TLS_CERT / TLS_KEY env vars).
//   - D-04: graceful shutdown via signal.NotifyContext.
//   - D-08: zerolog structured logging.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yourorg/cell-arch/internal/config"
	"github.com/yourorg/cell-arch/internal/sidecar/server"
	"github.com/yourorg/cell-arch/pkg/shutdown"
	taskpb "github.com/yourorg/cell-arch/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// 1. Initialise the structured logger.
	logger := initLogger()

	// 2. Root context cancelled on OS signals (D-04).
	ctx, stop := shutdown.Graceful(context.Background(), logger)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Fatal().Err(err).Msg("sidecar error")
	}
	logger.Info().Msg("sidecar stopped gracefully")
}

// run contains the main sidecar logic.
func run(ctx context.Context, logger zerolog.Logger) error {
	// 3. Load typed configuration (env vars + sidecar.yaml via plan 03).
	cfg, err := config.LoadSidecarConfig(ctx)
	if err != nil {
		return fmt.Errorf("load sidecar config: %w", err)
	}

	// 4. Build mTLS server credentials (D-15).
	creds, err := buildServerTLSCredentials(cfg)
	if err != nil {
		return fmt.Errorf("build mTLS credentials: %w", err)
	}

	// 5. Start gRPC listener.
	addr := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	// 6. Build gRPC server with mTLS.
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	// 7. Register TaskServiceServer (D-09, D-12).
	taskpb.RegisterTaskServiceServer(grpcServer, server.NewTaskServer(logger))

	logger.Info().Str("addr", addr).Msg("sidecar starting with mTLS")

	// 8. Serve in background; cancel on context cancellation.
	errCh := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc serve: %w", err)
		}
		close(errCh)
	}()

	// 9. Block until OS signal or serve error.
	select {
	case <-ctx.Done():
		logger.Info().Msg("shutdown signal received, draining connections")
		grpcServer.GracefulStop()
		return <-errCh
	case err := <-errCh:
		return err
	}
}

// buildServerTLSCredentials creates mTLS credentials from TLS_CERT and TLS_KEY env vars.
// The client CA cert is loaded from TLS_CA_CERT env var (optional; defaults to system pool).
func buildServerTLSCredentials(cfg *config.SidecarConfig) (credentials.TransportCredentials, error) {
	if cfg.TLSCert == "" || cfg.TLSKey == "" {
		return nil, fmt.Errorf("TLS_CERT and TLS_KEY environment variables are required for mTLS")
	}

	// Load server certificate and private key.
	cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
	if err != nil {
		return nil, fmt.Errorf("load TLS key pair: %w", err)
	}

	// Build the client CA pool for mutual authentication.
	clientCA := x509.NewCertPool()
	caPath := os.Getenv("TLS_CA_CERT")
	if caPath != "" {
		caCert, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("read CA cert %s: %w", caPath, err)
		}
		if !clientCA.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA cert from %s", caPath)
		}
	} else {
		// Fall back to the system CA pool; still enforces client cert.
		var errPool error
		clientCA, errPool = x509.SystemCertPool()
		if errPool != nil {
			return nil, fmt.Errorf("load system cert pool: %w", errPool)
		}
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCA,
		MinVersion:   tls.VersionTLS13,
	}
	return credentials.NewTLS(tlsCfg), nil
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

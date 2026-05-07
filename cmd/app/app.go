// Package main — app.go defines the application container.
// All dependencies are wired here via manual constructors (D-02).
// No init() side effects, no global state.
package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/yourorg/cell-arch/internal/config"
	"github.com/yourorg/cell-arch/internal/task/infrastructure/repo"
	"github.com/yourorg/cell-arch/internal/task/usecase"
)

// container holds all application dependencies.
// It is constructed once in main and passed to the HTTP/gRPC server in Phase 2.
type container struct {
	cfg         *config.AppConfig
	taskService *usecase.Service
	logger      zerolog.Logger
}

// newContainer builds and wires all application components.
// It is the single source of truth for the dependency graph.
func newContainer(ctx context.Context, logger zerolog.Logger) (*container, error) {
	// Load configuration from environment (boundary call per D-03).
	cfg, err := config.LoadAppConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("newContainer: load config: %w", err)
	}

	// Infrastructure layer: sidecar repository (no cloud SDKs here).
	taskRepo := repo.NewSidecarRepo(cfg.SidecarAddr, "aws", logger)

	// Use-case layer: business logic wired with repository.
	taskSvc := usecase.NewService(taskRepo, logger)

	return &container{
		cfg:         cfg,
		taskService: taskSvc,
		logger:      logger,
	}, nil
}

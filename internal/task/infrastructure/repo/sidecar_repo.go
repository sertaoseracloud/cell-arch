// Package repo provides infrastructure implementations of the task Repository
// interface. The SidecarRepo delegates all persistence calls to the local
// sidecar gRPC server — it contains NO cloud SDK imports.
package repo

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/yourorg/cell-arch/internal/task/entity"
)

// SidecarRepo implements entity.Repository by forwarding calls to the sidecar.
// The gRPC client is injected rather than constructed here to keep the
// constructor shallow (D-02).
type SidecarRepo struct {
	// addr is the sidecar gRPC address (e.g. "localhost:50051").
	addr   string
	cloud  string
	logger zerolog.Logger
}

// NewSidecarRepo constructs a SidecarRepo. The actual gRPC connection is
// established lazily on first call to allow the binary to start without
// a live sidecar (useful for testing).
func NewSidecarRepo(addr, cloud string, logger zerolog.Logger) *SidecarRepo {
	return &SidecarRepo{
		addr:   addr,
		cloud:  cloud,
		logger: logger,
	}
}

// Get retrieves a task from the sidecar.
// NOTE: Full gRPC wiring is implemented in Phase 2 (sidecar proxy plan).
func (r *SidecarRepo) Get(ctx context.Context, id string) (*entity.Task, error) {
	r.logger.Debug().Str("id", id).Str("cloud", r.cloud).Str("sidecar", r.addr).Msg("SidecarRepo.Get")
	return nil, fmt.Errorf("SidecarRepo.Get: not yet implemented (Phase 2): %w", errNotImplemented)
}

// Create persists a task via the sidecar.
func (r *SidecarRepo) Create(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	r.logger.Debug().Str("cloud", r.cloud).Msg("SidecarRepo.Create")
	return nil, fmt.Errorf("SidecarRepo.Create: not yet implemented (Phase 2): %w", errNotImplemented)
}

// Update modifies a task via the sidecar.
func (r *SidecarRepo) Update(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	r.logger.Debug().Str("id", t.ID).Str("cloud", r.cloud).Msg("SidecarRepo.Update")
	return nil, fmt.Errorf("SidecarRepo.Update: not yet implemented (Phase 2): %w", errNotImplemented)
}

// Delete removes a task via the sidecar.
func (r *SidecarRepo) Delete(ctx context.Context, id string) error {
	r.logger.Debug().Str("id", id).Str("cloud", r.cloud).Msg("SidecarRepo.Delete")
	return fmt.Errorf("SidecarRepo.Delete: not yet implemented (Phase 2): %w", errNotImplemented)
}

// List retrieves all tasks from the sidecar.
func (r *SidecarRepo) List(ctx context.Context) ([]*entity.Task, error) {
	r.logger.Debug().Str("cloud", r.cloud).Msg("SidecarRepo.List")
	return nil, fmt.Errorf("SidecarRepo.List: not yet implemented (Phase 2): %w", errNotImplemented)
}

// errNotImplemented is a sentinel used until Phase 2 wires the real gRPC client.
var errNotImplemented = fmt.Errorf("not implemented")

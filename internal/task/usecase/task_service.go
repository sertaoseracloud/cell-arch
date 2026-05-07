// Package usecase implements the business logic for the task feature.
// It depends only on the domain entity interfaces — no cloud SDKs.
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourorg/cell-arch/internal/task/entity"
)

// Service orchestrates task operations through the Repository interface.
// Dependencies are injected via the constructor (D-02).
type Service struct {
	repo   entity.Repository
	logger zerolog.Logger
}

// NewService wires a Service with its required dependencies.
// The caller (cmd/app/main.go) is responsible for providing concrete implementations.
func NewService(repo entity.Repository, logger zerolog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Get retrieves a task by ID. ctx is accepted as first argument per D-03.
func (s *Service) Get(ctx context.Context, id string) (*entity.Task, error) {
	s.logger.Debug().Str("task_id", id).Msg("Service.Get")
	task, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: get task %q: %w", id, err)
	}
	return task, nil
}

// Create persists a new task, setting timestamps and default status.
func (s *Service) Create(ctx context.Context, title, description string) (*entity.Task, error) {
	s.logger.Debug().Str("title", title).Msg("Service.Create")
	now := time.Now().UTC()
	t := &entity.Task{
		Title:       title,
		Description: description,
		Status:      entity.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	created, err := s.repo.Create(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("service: create task: %w", err)
	}
	return created, nil
}

// Update applies changes to an existing task, bumping its UpdatedAt timestamp.
func (s *Service) Update(ctx context.Context, id, title, description string, status entity.Status) (*entity.Task, error) {
	s.logger.Debug().Str("task_id", id).Msg("Service.Update")
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: get task before update %q: %w", id, err)
	}
	existing.Title = title
	existing.Description = description
	existing.Status = status
	existing.UpdatedAt = time.Now().UTC()

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("service: update task %q: %w", id, err)
	}
	return updated, nil
}

// Delete removes a task by ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	s.logger.Debug().Str("task_id", id).Msg("Service.Delete")
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service: delete task %q: %w", id, err)
	}
	return nil
}

// List returns all tasks.
func (s *Service) List(ctx context.Context) ([]*entity.Task, error) {
	s.logger.Debug().Msg("Service.List")
	tasks, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: list tasks: %w", err)
	}
	return tasks, nil
}

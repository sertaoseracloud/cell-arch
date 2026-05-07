// Package entity defines the core domain model for tasks.
// No external dependencies are allowed here — only stdlib.
package entity

import (
	"context"
	"time"
)

// Status represents the lifecycle state of a task.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Task is the core domain entity. It contains no cloud-SDK types.
type Task struct {
	ID          string
	Title       string
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Repository defines the persistence contract for tasks.
// Implementations live in internal/task/infrastructure/repo and talk to the sidecar.
type Repository interface {
	// Get retrieves a task by its ID.
	Get(ctx context.Context, id string) (*Task, error)
	// Create persists a new task and returns the created entity.
	Create(ctx context.Context, t *Task) (*Task, error)
	// Update modifies an existing task.
	Update(ctx context.Context, t *Task) (*Task, error)
	// Delete removes a task by its ID.
	Delete(ctx context.Context, id string) error
	// List returns all tasks.
	List(ctx context.Context) ([]*Task, error)
}

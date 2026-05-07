package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/cell-arch/internal/task/entity"
)

func TestStatus_Constants(t *testing.T) {
	tests := []struct {
		name   string
		status entity.Status
		want   string
	}{
		{
			name:   "pending status value",
			status: entity.StatusPending,
			want:   "pending",
		},
		{
			name:   "in_progress status value",
			status: entity.StatusInProgress,
			want:   "in_progress",
		},
		{
			name:   "done status value",
			status: entity.StatusDone,
			want:   "done",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, entity.Status(tc.want), tc.status)
		})
	}
}

func TestTask_Fields(t *testing.T) {
	t.Run("task struct holds all fields", func(t *testing.T) {
		task := entity.Task{
			ID:          "task-1",
			Title:       "Write tests",
			Description: "Implement unit tests for domain layer",
			Status:      entity.StatusPending,
		}

		assert.Equal(t, "task-1", task.ID)
		assert.Equal(t, "Write tests", task.Title)
		assert.Equal(t, "Implement unit tests for domain layer", task.Description)
		assert.Equal(t, entity.StatusPending, task.Status)
		assert.True(t, task.CreatedAt.IsZero(), "CreatedAt should be zero-valued when not set")
		assert.True(t, task.UpdatedAt.IsZero(), "UpdatedAt should be zero-valued when not set")
	})

	t.Run("task status transitions", func(t *testing.T) {
		task := &entity.Task{
			ID:     "task-2",
			Status: entity.StatusPending,
		}

		task.Status = entity.StatusInProgress
		assert.Equal(t, entity.StatusInProgress, task.Status)

		task.Status = entity.StatusDone
		assert.Equal(t, entity.StatusDone, task.Status)
	})
}

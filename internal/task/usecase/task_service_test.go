package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/cell-arch/internal/task/entity"
	"github.com/yourorg/cell-arch/internal/task/usecase"
)

func newNopLogger() zerolog.Logger {
	return zerolog.Nop()
}

func TestService_Get(t *testing.T) {
	tests := []struct {
		name        string
		taskID      string
		repoTask    *entity.Task
		repoErr     error
		wantTask    *entity.Task
		wantErr     bool
		errContains string
	}{
		{
			name:   "successfully retrieves task",
			taskID: "task-1",
			repoTask: &entity.Task{
				ID:     "task-1",
				Title:  "Test task",
				Status: entity.StatusPending,
			},
			repoErr:  nil,
			wantTask: &entity.Task{ID: "task-1", Title: "Test task", Status: entity.StatusPending},
			wantErr:  false,
		},
		{
			name:        "returns error when repo fails",
			taskID:      "task-missing",
			repoTask:    nil,
			repoErr:     errors.New("not found"),
			wantTask:    nil,
			wantErr:     true,
			errContains: "task-missing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockRepository)
			repo.On("Get", mock.Anything, tc.taskID).Return(tc.repoTask, tc.repoErr)

			svc := usecase.NewService(repo, newNopLogger())
			got, err := svc.Get(context.Background(), tc.taskID)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantTask, got)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestService_Create(t *testing.T) {
	t.Run("successfully creates task", func(t *testing.T) {
		returnedTask := &entity.Task{
			ID:          "generated-id",
			Title:       "New task",
			Description: "Description",
			Status:      entity.StatusPending,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		repo := new(mockRepository)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(task *entity.Task) bool {
			return task.Title == "New task" &&
				task.Description == "Description" &&
				task.Status == entity.StatusPending &&
				!task.CreatedAt.IsZero()
		})).Return(returnedTask, nil)

		svc := usecase.NewService(repo, newNopLogger())
		got, err := svc.Create(context.Background(), "New task", "Description")

		require.NoError(t, err)
		assert.Equal(t, "New task", got.Title)
		assert.Equal(t, entity.StatusPending, got.Status)
		assert.False(t, got.CreatedAt.IsZero())
		repo.AssertExpectations(t)
	})

	t.Run("returns error when repo fails", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Task")).
			Return((*entity.Task)(nil), errors.New("storage error"))

		svc := usecase.NewService(repo, newNopLogger())
		got, err := svc.Create(context.Background(), "Bad task", "Desc")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "create task")
		assert.Nil(t, got)
		repo.AssertExpectations(t)
	})
}

func TestService_Update(t *testing.T) {
	existingTask := &entity.Task{
		ID:          "task-1",
		Title:       "Old title",
		Description: "Old desc",
		Status:      entity.StatusPending,
		CreatedAt:   time.Now().Add(-time.Hour),
	}

	tests := []struct {
		name        string
		taskID      string
		newTitle    string
		newDesc     string
		newStatus   entity.Status
		getResp     *entity.Task
		getErr      error
		updateResp  *entity.Task
		updateErr   error
		wantErr     bool
		errContains string
	}{
		{
			name:      "successfully updates task",
			taskID:    "task-1",
			newTitle:  "New title",
			newDesc:   "New desc",
			newStatus: entity.StatusInProgress,
			getResp:   existingTask,
			getErr:    nil,
			updateResp: &entity.Task{
				ID:     "task-1",
				Title:  "New title",
				Status: entity.StatusInProgress,
			},
			updateErr: nil,
			wantErr:   false,
		},
		{
			name:        "returns error when get fails",
			taskID:      "task-missing",
			newTitle:    "X",
			newDesc:     "Y",
			newStatus:   entity.StatusDone,
			getResp:     nil,
			getErr:      errors.New("not found"),
			updateResp:  nil,
			updateErr:   nil,
			wantErr:     true,
			errContains: "task-missing",
		},
		{
			name:        "returns error when update fails",
			taskID:      "task-1",
			newTitle:    "New title",
			newDesc:     "New desc",
			newStatus:   entity.StatusDone,
			getResp:     existingTask,
			getErr:      nil,
			updateResp:  nil,
			updateErr:   errors.New("storage error"),
			wantErr:     true,
			errContains: "task-1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockRepository)
			repo.On("Get", mock.Anything, tc.taskID).Return(tc.getResp, tc.getErr)

			if tc.getErr == nil {
				repo.On("Update", mock.Anything, mock.MatchedBy(func(t *entity.Task) bool {
					return t.ID == tc.taskID && t.Title == tc.newTitle
				})).Return(tc.updateResp, tc.updateErr)
			}

			svc := usecase.NewService(repo, newNopLogger())
			got, err := svc.Update(context.Background(), tc.taskID, tc.newTitle, tc.newDesc, tc.newStatus)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.newTitle, got.Title)
				assert.Equal(t, tc.newStatus, got.Status)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		taskID      string
		repoErr     error
		wantErr     bool
		errContains string
	}{
		{
			name:    "successfully deletes task",
			taskID:  "task-1",
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns error when repo fails",
			taskID:      "task-missing",
			repoErr:     errors.New("not found"),
			wantErr:     true,
			errContains: "task-missing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockRepository)
			repo.On("Delete", mock.Anything, tc.taskID).Return(tc.repoErr)

			svc := usecase.NewService(repo, newNopLogger())
			err := svc.Delete(context.Background(), tc.taskID)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name      string
		repoTasks []*entity.Task
		repoErr   error
		wantCount int
		wantErr   bool
	}{
		{
			name: "returns list of tasks",
			repoTasks: []*entity.Task{
				{ID: "task-1", Title: "Task 1", Status: entity.StatusPending},
				{ID: "task-2", Title: "Task 2", Status: entity.StatusDone},
			},
			repoErr:   nil,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "returns empty list",
			repoTasks: []*entity.Task{},
			repoErr:   nil,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "returns error when repo fails",
			repoTasks: nil,
			repoErr:   errors.New("storage error"),
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockRepository)
			repo.On("List", mock.Anything).Return(tc.repoTasks, tc.repoErr)

			svc := usecase.NewService(repo, newNopLogger())
			got, err := svc.List(context.Background())

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tc.wantCount)
			}
			repo.AssertExpectations(t)
		})
	}
}

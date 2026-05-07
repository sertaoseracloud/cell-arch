// Package server_test contains unit tests for TaskServer cloud routing and error mapping.
// A mock CloudBackend is used so no real cloud credentials are required.
package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	taskpb "github.com/yourorg/cell-arch/proto"

	"github.com/yourorg/cell-arch/internal/sidecar/server"
)

// ---- Mock CloudBackend --------------------------------------------------

type mockBackend struct {
	mock.Mock
}

func (m *mockBackend) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockBackend) Create(ctx context.Context, item map[string]interface{}) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *mockBackend) Query(ctx context.Context) ([]map[string]interface{}, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *mockBackend) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ---- helpers ------------------------------------------------------------

func newServerWithMock(t *testing.T) (*server.TaskServer, *mockBackend) {
	t.Helper()
	logger := zerolog.Nop()
	s := server.NewTaskServer(logger)
	b := &mockBackend{}
	s.RegisterBackend("aws", b)
	return s, b
}

func itemMap(id, title string) map[string]interface{} {
	return map[string]interface{}{
		"id":          id,
		"title":       title,
		"description": "",
		"status":      "pending",
		"created_at":  "2026-01-01T00:00:00Z",
		"updated_at":  "2026-01-01T00:00:00Z",
	}
}

// ---- HealthCheck --------------------------------------------------------

func TestHealthCheck_ReturnsServing(t *testing.T) {
	s := server.NewTaskServer(zerolog.Nop())
	// Register a backend so health check returns SERVING
	b := &mockBackend{}
	s.RegisterBackend("aws", b)
	resp, err := s.HealthCheck(context.Background(), &taskpb.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, taskpb.ServingStatus_SERVING, resp.Status)
}

func TestHealthCheck_NoBackends(t *testing.T) {
	s := server.NewTaskServer(zerolog.Nop())
	// No backends registered
	resp, err := s.HealthCheck(context.Background(), &taskpb.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, taskpb.ServingStatus_NOT_SERVING, resp.Status)
}

// ---- Unknown cloud backend (InvalidArgument) ----------------------------

func TestGetTask_UnknownCloud(t *testing.T) {
	s := server.NewTaskServer(zerolog.Nop())
	_, err := s.GetTask(context.Background(), &taskpb.GetTaskRequest{TaskId: "id", Cloud: "gcp"})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCreateTask_UnknownCloud(t *testing.T) {
	s := server.NewTaskServer(zerolog.Nop())
	_, err := s.CreateTask(context.Background(), &taskpb.CreateTaskRequest{Title: "t", Cloud: "gcp"})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestQueryTasks_UnknownCloud(t *testing.T) {
	s := server.NewTaskServer(zerolog.Nop())
	_, err := s.QueryTasks(context.Background(), &taskpb.QueryTasksRequest{Cloud: "gcp"})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestDeleteTask_UnknownCloud(t *testing.T) {
	s := server.NewTaskServer(zerolog.Nop())
	_, err := s.DeleteTask(context.Background(), &taskpb.DeleteTaskRequest{TaskId: "id", Cloud: "gcp"})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

// ---- GetTask ------------------------------------------------------------

func TestGetTask_Success(t *testing.T) {
	s, b := newServerWithMock(t)
	b.On("Get", mock.Anything, "task-1").Return(itemMap("task-1", "My Task"), nil)

	resp, err := s.GetTask(context.Background(), &taskpb.GetTaskRequest{TaskId: "task-1", Cloud: "aws"})
	require.NoError(t, err)
	require.NotNil(t, resp.Task)
	assert.Equal(t, "task-1", resp.Task.Id)
	assert.Equal(t, "My Task", resp.Task.Title)
	b.AssertExpectations(t)
}

func TestGetTask_NotFound(t *testing.T) {
	s, b := newServerWithMock(t)
	b.On("Get", mock.Anything, "missing").Return(nil, server.ErrItemNotFound)

	_, err := s.GetTask(context.Background(), &taskpb.GetTaskRequest{TaskId: "missing", Cloud: "aws"})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
	b.AssertExpectations(t)
}

func TestGetTask_InternalError(t *testing.T) {
	s, b := newServerWithMock(t)
	b.On("Get", mock.Anything, "task-1").Return(nil, errors.New("connection refused"))

	_, err := s.GetTask(context.Background(), &taskpb.GetTaskRequest{TaskId: "task-1", Cloud: "aws"})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	b.AssertExpectations(t)
}

// ---- CreateTask ---------------------------------------------------------

func TestCreateTask_Success(t *testing.T) {
	s, b := newServerWithMock(t)
	b.On("Create", mock.Anything, mock.AnythingOfType("map[string]interface {}")).Return(nil)

	resp, err := s.CreateTask(context.Background(), &taskpb.CreateTaskRequest{
		Title:       "New Task",
		Description: "desc",
		Cloud:       "aws",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Task)
	assert.Equal(t, "New Task", resp.Task.Title)
	assert.NotEmpty(t, resp.Task.Id)
	b.AssertExpectations(t)
}

func TestCreateTask_InternalError(t *testing.T) {
	s, b := newServerWithMock(t)
	b.On("Create", mock.Anything, mock.AnythingOfType("map[string]interface {}")).Return(errors.New("write error"))

	_, err := s.CreateTask(context.Background(), &taskpb.CreateTaskRequest{Title: "t", Cloud: "aws"})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	b.AssertExpectations(t)
}

// ---- QueryTasks ---------------------------------------------------------

func TestQueryTasks_Success(t *testing.T) {
	s, b := newServerWithMock(t)
	items := []map[string]interface{}{
		itemMap("task-1", "Task One"),
		itemMap("task-2", "Task Two"),
	}
	b.On("Query", mock.Anything).Return(items, nil)

	resp, err := s.QueryTasks(context.Background(), &taskpb.QueryTasksRequest{Cloud: "aws"})
	require.NoError(t, err)
	assert.Len(t, resp.Tasks, 2)
	b.AssertExpectations(t)
}

func TestQueryTasks_Empty(t *testing.T) {
	s, b := newServerWithMock(t)
	b.On("Query", mock.Anything).Return([]map[string]interface{}{}, nil)

	resp, err := s.QueryTasks(context.Background(), &taskpb.QueryTasksRequest{Cloud: "aws"})
	require.NoError(t, err)
	assert.Empty(t, resp.Tasks)
	b.AssertExpectations(t)
}

func TestQueryTasks_InternalError(t *testing.T) {
	s, b := newServerWithMock(t)
	b.On("Query", mock.Anything).Return(nil, errors.New("scan failed"))

	_, err := s.QueryTasks(context.Background(), &taskpb.QueryTasksRequest{Cloud: "aws"})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	b.AssertExpectations(t)
}

// ---- DeleteTask ---------------------------------------------------------

func TestDeleteTask_Success(t *testing.T) {
	s, b := newServerWithMock(t)
	b.On("Delete", mock.Anything, "task-1").Return(nil)

	resp, err := s.DeleteTask(context.Background(), &taskpb.DeleteTaskRequest{TaskId: "task-1", Cloud: "aws"})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	b.AssertExpectations(t)
}

func TestDeleteTask_InternalError(t *testing.T) {
	s, b := newServerWithMock(t)
	b.On("Delete", mock.Anything, "task-1").Return(errors.New("delete failed"))

	_, err := s.DeleteTask(context.Background(), &taskpb.DeleteTaskRequest{TaskId: "task-1", Cloud: "aws"})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	b.AssertExpectations(t)
}

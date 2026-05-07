// Package server implements the TaskServiceServer gRPC interface for the sidecar.
// Cloud SDK calls are delegated to backends registered by cloud name (D-12).
// This layer is responsible only for: request validation, cloud routing,
// attribute mapping, error mapping to gRPC status codes (D-13), and returning typed responses.
package server

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	taskpb "github.com/yourorg/cell-arch/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CloudBackend is implemented by each cloud adapter (AWS DynamoDB, Azure CosmosDB).
// All methods accept context.Context as the first argument (D-03).
type CloudBackend interface {
	Get(ctx context.Context, id string) (map[string]interface{}, error)
	Create(ctx context.Context, item map[string]interface{}) error
	Query(ctx context.Context) ([]map[string]interface{}, error)
	Delete(ctx context.Context, id string) error
}

// ErrItemNotFound must be checked via errors.Is to map to codes.NotFound.
// Cloud adapters must wrap their own sentinel with this value when an item is missing.
var ErrItemNotFound = errors.New("item not found")

// TaskServer implements taskpb.TaskServiceServer.
// Cloud backends are registered by name (e.g., "aws", "azure") — routes per request field.
type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	logger   zerolog.Logger
	backends map[string]CloudBackend
}

// NewTaskServer constructs a TaskServer with the given logger (D-02).
// Call RegisterBackend to add cloud adapters after construction.
func NewTaskServer(logger zerolog.Logger) *TaskServer {
	return &TaskServer{
		logger:   logger,
		backends: make(map[string]CloudBackend),
	}
}

// RegisterBackend registers a CloudBackend under the given cloud key (e.g., "aws").
func (s *TaskServer) RegisterBackend(cloud string, backend CloudBackend) {
	s.backends[cloud] = backend
}

// backend resolves the requested cloud backend or returns a gRPC status error.
func (s *TaskServer) backend(cloud string) (CloudBackend, error) {
	b, ok := s.backends[cloud]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unknown cloud backend %q", cloud)
	}
	return b, nil
}

// mapToTask converts a generic map to a proto Task.
func mapToTask(item map[string]interface{}) *taskpb.Task {
	str := func(key string) string {
		if v, ok := item[key].(string); ok {
			return v
		}
		return ""
	}
	return &taskpb.Task{
		Id:          str("id"),
		Title:       str("title"),
		Description: str("description"),
		Status:      str("status"),
		CreatedAt:   str("created_at"),
		UpdatedAt:   str("updated_at"),
	}
}

// mapError translates cloud adapter errors to gRPC status errors (D-13).
func mapError(err error) error {
	if errors.Is(err, ErrItemNotFound) {
		return status.Errorf(codes.NotFound, "task not found")
	}
	return status.Errorf(codes.Internal, "backend error: %v", err)
}

// HealthCheck returns SERVING when the sidecar is up (D-14).
// NOTE: This is a lightweight check — it only verifies that backends are registered.
// Per D-14, no cloud connectivity probes are performed on each health check.
// For production, consider periodic background probes for each backend.
func (s *TaskServer) HealthCheck(ctx context.Context, _ *taskpb.HealthCheckRequest) (*taskpb.HealthCheckResponse, error) {
	s.logger.Debug().Msg("HealthCheck called")
	awsOk := s.backends["aws"] != nil
	azureOk := s.backends["azure"] != nil

	if !awsOk && !azureOk {
		return &taskpb.HealthCheckResponse{
			Status: taskpb.ServingStatus_NOT_SERVING,
		}, nil
	}
	return &taskpb.HealthCheckResponse{
		Status: taskpb.ServingStatus_SERVING,
	}, nil
}

// GetTask retrieves a task by ID from the requested cloud backend (D-12).
func (s *TaskServer) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.GetTaskResponse, error) {
	s.logger.Debug().Str("task_id", req.TaskId).Str("cloud", req.Cloud).Msg("GetTask called")
	b, err := s.backend(req.Cloud)
	if err != nil {
		return nil, err
	}
	item, err := b.Get(ctx, req.TaskId)
	if err != nil {
		return nil, mapError(err)
	}
	return &taskpb.GetTaskResponse{Task: mapToTask(item)}, nil
}

// CreateTask persists a new task to the requested cloud backend (D-12).
func (s *TaskServer) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.CreateTaskResponse, error) {
	s.logger.Debug().Str("cloud", req.Cloud).Msg("CreateTask called")
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	b, err := s.backend(req.Cloud)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	id := uuid.New().String()
	item := map[string]interface{}{
		"id":          id,
		"title":       req.Title,
		"description": req.Description,
		"status":      "pending",
		"created_at":  now,
		"updated_at":  now,
	}
	if err := b.Create(ctx, item); err != nil {
		return nil, mapError(err)
	}
	return &taskpb.CreateTaskResponse{Task: mapToTask(item)}, nil
}

// QueryTasks returns all tasks from the requested cloud backend (D-12).
func (s *TaskServer) QueryTasks(ctx context.Context, req *taskpb.QueryTasksRequest) (*taskpb.QueryTasksResponse, error) {
	s.logger.Debug().Str("cloud", req.Cloud).Msg("QueryTasks called")
	b, err := s.backend(req.Cloud)
	if err != nil {
		return nil, err
	}
	items, err := b.Query(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	tasks := make([]*taskpb.Task, 0, len(items))
	for _, it := range items {
		tasks = append(tasks, mapToTask(it))
	}
	return &taskpb.QueryTasksResponse{Tasks: tasks}, nil
}

// DeleteTask removes a task from the requested cloud backend (D-12).
func (s *TaskServer) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*taskpb.DeleteTaskResponse, error) {
	s.logger.Debug().Str("task_id", req.TaskId).Str("cloud", req.Cloud).Msg("DeleteTask called")
	b, err := s.backend(req.Cloud)
	if err != nil {
		return nil, err
	}
	if err := b.Delete(ctx, req.TaskId); err != nil {
		return nil, mapError(err)
	}
	return &taskpb.DeleteTaskResponse{Success: true}, nil
}

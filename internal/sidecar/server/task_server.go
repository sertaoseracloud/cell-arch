// Package server implements the TaskServiceServer gRPC interface for the sidecar.
// Cloud SDK calls are delegated to the aws and azure sub-packages.
// This layer is responsible only for: request validation, cloud routing (D-12),
// error mapping to gRPC status codes (D-13), and returning typed responses.
package server

import (
	"context"

	"github.com/rs/zerolog"
	taskpb "github.com/yourorg/cell-arch/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TaskServer implements taskpb.TaskServiceServer.
// Phase 2 plan 01 provides the scaffold — cloud SDK wiring follows in subsequent plans.
type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	logger zerolog.Logger
}

// NewTaskServer constructs a TaskServer with the given logger (D-02).
func NewTaskServer(logger zerolog.Logger) *TaskServer {
	return &TaskServer{logger: logger}
}

// HealthCheck returns SERVING when the sidecar is up.
// Full cloud connectivity probing will be added in plan 02 (D-14).
func (s *TaskServer) HealthCheck(ctx context.Context, _ *taskpb.HealthCheckRequest) (*taskpb.HealthCheckResponse, error) {
	s.logger.Debug().Msg("HealthCheck called")
	return &taskpb.HealthCheckResponse{
		Status: taskpb.ServingStatus_SERVING,
	}, nil
}

// GetTask stubs the retrieval path — real implementation in plan 02.
func (s *TaskServer) GetTask(_ context.Context, req *taskpb.GetTaskRequest) (*taskpb.GetTaskResponse, error) {
	s.logger.Debug().Str("task_id", req.TaskId).Str("cloud", req.Cloud).Msg("GetTask called")
	return nil, status.Errorf(codes.Unimplemented, "GetTask: cloud backend not yet wired (Phase 2 plan 02)")
}

// CreateTask stubs the create path — real implementation in plan 02.
func (s *TaskServer) CreateTask(_ context.Context, req *taskpb.CreateTaskRequest) (*taskpb.CreateTaskResponse, error) {
	s.logger.Debug().Str("cloud", req.Cloud).Msg("CreateTask called")
	return nil, status.Errorf(codes.Unimplemented, "CreateTask: cloud backend not yet wired (Phase 2 plan 02)")
}

// QueryTasks stubs the list path — real implementation in plan 02.
func (s *TaskServer) QueryTasks(_ context.Context, req *taskpb.QueryTasksRequest) (*taskpb.QueryTasksResponse, error) {
	s.logger.Debug().Str("cloud", req.Cloud).Msg("QueryTasks called")
	return nil, status.Errorf(codes.Unimplemented, "QueryTasks: cloud backend not yet wired (Phase 2 plan 02)")
}

// DeleteTask stubs the delete path — real implementation in plan 02.
func (s *TaskServer) DeleteTask(_ context.Context, req *taskpb.DeleteTaskRequest) (*taskpb.DeleteTaskResponse, error) {
	s.logger.Debug().Str("task_id", req.TaskId).Str("cloud", req.Cloud).Msg("DeleteTask called")
	return nil, status.Errorf(codes.Unimplemented, "DeleteTask: cloud backend not yet wired (Phase 2 plan 02)")
}

// Code generated manually (protoc not available in build environment).
// Source: proto/task.proto
// Do not edit — regenerate with:
//   protoc --go_out=. --go-grpc_out=. proto/task.proto

package proto

// ServingStatus mirrors grpc.health.v1 semantics.
type ServingStatus int32

const (
	ServingStatus_UNKNOWN     ServingStatus = 0
	ServingStatus_SERVING     ServingStatus = 1
	ServingStatus_NOT_SERVING ServingStatus = 2
)

// Task is the wire representation of the domain Task entity.
type Task struct {
	Id          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// GetTaskRequest is sent by the client to retrieve a task by ID.
type GetTaskRequest struct {
	TaskId string `json:"task_id,omitempty"`
	// Cloud selects the backend provider: "aws" or "azure".
	Cloud string `json:"cloud,omitempty"`
}

// GetTaskResponse is the server's reply to GetTaskRequest.
type GetTaskResponse struct {
	Task *Task `json:"task,omitempty"`
}

// CreateTaskRequest is sent by the client to create a new task.
type CreateTaskRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	// Cloud selects the backend provider: "aws" or "azure".
	Cloud string `json:"cloud,omitempty"`
}

// CreateTaskResponse is the server's reply to CreateTaskRequest.
type CreateTaskResponse struct {
	Task *Task `json:"task,omitempty"`
}

// QueryTasksRequest is sent by the client to list all tasks.
type QueryTasksRequest struct {
	// Cloud selects the backend provider: "aws" or "azure".
	Cloud string `json:"cloud,omitempty"`
}

// QueryTasksResponse is the server's reply to QueryTasksRequest.
type QueryTasksResponse struct {
	Tasks []*Task `json:"tasks,omitempty"`
}

// DeleteTaskRequest is sent by the client to delete a task by ID.
type DeleteTaskRequest struct {
	TaskId string `json:"task_id,omitempty"`
	// Cloud selects the backend provider: "aws" or "azure".
	Cloud string `json:"cloud,omitempty"`
}

// DeleteTaskResponse is the server's reply to DeleteTaskRequest.
type DeleteTaskResponse struct {
	Success bool `json:"success,omitempty"`
}

// HealthCheckRequest carries no fields.
type HealthCheckRequest struct{}

// HealthCheckResponse carries the serving status of the sidecar.
type HealthCheckResponse struct {
	Status ServingStatus `json:"status,omitempty"`
}

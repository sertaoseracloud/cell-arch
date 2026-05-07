package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapAWSError maps AWS SDK errors to appropriate gRPC status codes.
func MapAWSError(err error) error {
	if err == nil {
		return nil
	}
	// Use fmt.Sprintf to inspect the error string for specific AWS error types.
	errStr := err.Error()
	switch {
	case contains(errStr, "ResourceNotFoundException"):
		return status.Errorf(codes.NotFound, "resource not found: %v", err)
	case contains(errStr, "ValidationException"):
		return status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	case contains(errStr, "ProvisionedThroughputExceededException"):
		return status.Errorf(codes.ResourceExhausted, "throughput exceeded: %v", err)
	case contains(errStr, "ConditionalCheckFailedException"):
		return status.Errorf(codes.FailedPrecondition, "conditional check failed: %v", err)
	default:
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}
}

// MapAzureError maps Azure SDK errors to appropriate gRPC status codes.
func MapAzureError(err error) error {
	if err == nil {
		return nil
	}
	errStr := err.Error()
	switch {
	case contains(errStr, "404") || contains(errStr, "Not Found") || contains(errStr, "ResourceNotFound"):
		return status.Errorf(codes.NotFound, "resource not found: %v", err)
	case contains(errStr, "400") || contains(errStr, "Bad Request") || contains(errStr, "invalid"):
		return status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	case contains(errStr, "429") || contains(errStr, "Throttling") || contains(errStr, "Too Many Requests"):
		return status.Errorf(codes.ResourceExhausted, "throttled: %v", err)
	case contains(errStr, "401") || contains(errStr, "Unauthorized"):
		return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
	default:
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}
}

// contains checks if substr exists in s.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstr(s, substr))
}

// findSubstr is a simple substring search.
func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

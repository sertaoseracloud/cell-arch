package errors

import (
	"errors"
	"strings"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dynamodbclient "github.com/yourorg/cell-arch/internal/sidecar/aws"
)

// MapAWSError maps AWS SDK errors to appropriate gRPC status codes.
// It first checks for the application-level ErrItemNotFound sentinel, then
// uses errors.As to match typed AWS SDK errors before falling back to Internal.
func MapAWSError(err error) error {
	if err == nil {
		return nil
	}

	// Application-level sentinel: item not found in DynamoDB.
	if errors.Is(err, dynamodbclient.ErrItemNotFound) {
		return status.Errorf(codes.NotFound, "item not found")
	}

	// AWS SDK typed errors — use errors.As for reliable classification.
	var notFound *dynamodbtypes.ResourceNotFoundException
	if errors.As(err, &notFound) {
		return status.Errorf(codes.NotFound, "resource not found: %v", err)
	}

	var throughputEx *dynamodbtypes.ProvisionedThroughputExceededException
	if errors.As(err, &throughputEx) {
		return status.Errorf(codes.ResourceExhausted, "throughput exceeded: %v", err)
	}

	var throttleEx *dynamodbtypes.ThrottlingException
	if errors.As(err, &throttleEx) {
		return status.Errorf(codes.ResourceExhausted, "throttled: %v", err)
	}

	var condCheckEx *dynamodbtypes.ConditionalCheckFailedException
	if errors.As(err, &condCheckEx) {
		return status.Errorf(codes.FailedPrecondition, "conditional check failed: %v", err)
	}

	// ValidationException surfaces as a generic API error (no dedicated type in the SDK).
	// Fall back to string matching on the error message.
	if strings.Contains(err.Error(), "ValidationException") {
		return status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	}

	return status.Errorf(codes.Internal, "internal error: %v", err)
}

// MapAzureError maps Azure SDK errors to appropriate gRPC status codes.
// String-based matching is used because the Azure SDK for Go does not expose
// fine-grained typed error types equivalent to the AWS SDK.
func MapAzureError(err error) error {
	if err == nil {
		return nil
	}
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "404") || strings.Contains(errStr, "Not Found") || strings.Contains(errStr, "ResourceNotFound"):
		return status.Errorf(codes.NotFound, "resource not found: %v", err)
	case strings.Contains(errStr, "400") || strings.Contains(errStr, "Bad Request"):
		return status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	case strings.Contains(errStr, "429") || strings.Contains(errStr, "Throttling") || strings.Contains(errStr, "Too Many Requests"):
		return status.Errorf(codes.ResourceExhausted, "throttled: %v", err)
	case strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized"):
		return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
	default:
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}
}

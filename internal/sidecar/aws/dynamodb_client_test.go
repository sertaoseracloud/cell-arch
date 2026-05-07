// Package aws_test contains unit tests for the DynamoDB client.
// All tests mock the DynamoDBAPI interface — no real AWS calls are made.
package aws_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	awssidecar "github.com/yourorg/cell-arch/internal/sidecar/aws"
)

// ---- Mock ---------------------------------------------------------------

// mockDynamoDBAPI is a testify mock implementing DynamoDBAPI.
type mockDynamoDBAPI struct {
	mock.Mock
}

func (m *mockDynamoDBAPI) GetItem(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *mockDynamoDBAPI) PutItem(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *mockDynamoDBAPI) Scan(ctx context.Context, input *dynamodb.ScanInput, opts ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.ScanOutput), args.Error(1)
}

func (m *mockDynamoDBAPI) DeleteItem(ctx context.Context, input *dynamodb.DeleteItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

// ---- helpers ------------------------------------------------------------

func newTestClient(t *testing.T, m *mockDynamoDBAPI) *awssidecar.DynamoDBClient {
	t.Helper()
	logger := zerolog.Nop()
	return awssidecar.NewDynamoDBClientFromAPI(m, "test-table", logger)
}

func avItem(id, title string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":     &types.AttributeValueMemberS{Value: id},
		"title":  &types.AttributeValueMemberS{Value: title},
	}
}

func ifaceItem(id, title string) map[string]interface{} {
	return map[string]interface{}{
		"id":    id,
		"title": title,
	}
}

// ---- Get ----------------------------------------------------------------

func TestGet_Success(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	item := avItem("task-1", "My Task")
	m.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(&dynamodb.GetItemOutput{Item: item}, nil)

	got, err := client.Get(context.Background(), "task-1")
	require.NoError(t, err)
	assert.Equal(t, "task-1", got["id"])
	assert.Equal(t, "My Task", got["title"])
	m.AssertExpectations(t)
}

func TestGet_NotFound(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	m.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(&dynamodb.GetItemOutput{Item: nil}, nil)

	_, err := client.Get(context.Background(), "missing-id")
	require.Error(t, err)
	assert.True(t, errors.Is(err, awssidecar.ErrItemNotFound), "expected ErrItemNotFound, got: %v", err)
	m.AssertExpectations(t)
}

func TestGet_APIError(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	apiErr := errors.New("connection refused")
	m.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(nil, apiErr)

	_, err := client.Get(context.Background(), "task-1")
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to get item")
	m.AssertExpectations(t)
}

// ---- Create -------------------------------------------------------------

func TestCreate_Success(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	item := ifaceItem("task-2", "New Task")
	m.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput")).
		Return(&dynamodb.PutItemOutput{}, nil)

	err := client.Create(context.Background(), item)
	require.NoError(t, err)
	m.AssertExpectations(t)
}

func TestCreate_APIError(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	apiErr := errors.New("provisioned throughput exceeded")
	m.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput")).
		Return(nil, apiErr)

	item := ifaceItem("task-3", "Task")
	err := client.Create(context.Background(), item)
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to put item")
	m.AssertExpectations(t)
}

// ---- Query (Scan) -------------------------------------------------------

func TestQuery_Success(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	items := []map[string]types.AttributeValue{
		avItem("task-1", "Task One"),
		avItem("task-2", "Task Two"),
	}
	m.On("Scan", mock.Anything, mock.AnythingOfType("*dynamodb.ScanInput")).
		Return(&dynamodb.ScanOutput{Items: items}, nil)

	got, err := client.Query(context.Background())
	require.NoError(t, err)
	assert.Len(t, got, 2)
	m.AssertExpectations(t)
}

func TestQuery_Empty(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	m.On("Scan", mock.Anything, mock.AnythingOfType("*dynamodb.ScanInput")).
		Return(&dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{}}, nil)

	got, err := client.Query(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
	m.AssertExpectations(t)
}

func TestQuery_APIError(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	apiErr := errors.New("table not found")
	m.On("Scan", mock.Anything, mock.AnythingOfType("*dynamodb.ScanInput")).
		Return(nil, apiErr)

	_, err := client.Query(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to scan items")
	m.AssertExpectations(t)
}

// ---- Delete -------------------------------------------------------------

func TestDelete_Success(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	m.On("DeleteItem", mock.Anything, mock.AnythingOfType("*dynamodb.DeleteItemInput")).
		Return(&dynamodb.DeleteItemOutput{}, nil)

	err := client.Delete(context.Background(), "task-1")
	require.NoError(t, err)
	m.AssertExpectations(t)
}

func TestDelete_APIError(t *testing.T) {
	m := &mockDynamoDBAPI{}
	client := newTestClient(t, m)

	apiErr := errors.New("item locked")
	m.On("DeleteItem", mock.Anything, mock.AnythingOfType("*dynamodb.DeleteItemInput")).
		Return(nil, apiErr)

	err := client.Delete(context.Background(), "task-1")
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to delete item")
	m.AssertExpectations(t)
}

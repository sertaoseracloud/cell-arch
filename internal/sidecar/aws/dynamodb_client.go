// Package aws provides the AWS DynamoDB adapter for the sidecar proxy.
// Authentication uses IRSA (D-10) — no static credentials are accepted.
// All operations accept context.Context as the first argument (D-03).
package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

// ErrItemNotFound is returned by Get when the requested item does not exist in DynamoDB.
var ErrItemNotFound = errors.New("item not found")

// DynamoDBAPI is the subset of dynamodb.Client methods used by DynamoDBClient.
// Declaring this interface enables testify mocking without the AWS SDK wire layer.
type DynamoDBAPI interface {
	GetItem(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Scan(ctx context.Context, input *dynamodb.ScanInput, opts ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	DeleteItem(ctx context.Context, input *dynamodb.DeleteItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

// DynamoDBClient wraps the AWS DynamoDB SDK with project-level error mapping and logging.
type DynamoDBClient struct {
	api       DynamoDBAPI
	tableName string
	logger    zerolog.Logger
}

// avMapToInterface converts AWS DynamoDB attribute value map to a generic map[string]interface{}.
func avMapToInterface(av map[string]types.AttributeValue) map[string]interface{} {
	out := make(map[string]interface{}, len(av))
	for k, v := range av {
		switch val := v.(type) {
		case *types.AttributeValueMemberS:
			out[k] = val.Value
		case *types.AttributeValueMemberN:
			out[k] = val.Value
		case *types.AttributeValueMemberBOOL:
			out[k] = val.Value
		case *types.AttributeValueMemberM:
			out[k] = avMapToInterface(val.Value)
		default:
			out[k] = nil
		}
	}
	return out
}

// interfaceMapToAttributeValue converts a generic map[string]interface{} to DynamoDB attribute values.
func interfaceMapToAttributeValue(m map[string]interface{}) map[string]types.AttributeValue {
	av := make(map[string]types.AttributeValue, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case string:
			av[k] = &types.AttributeValueMemberS{Value: val}
		case int, int32, int64, float32, float64:
			av[k] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%v", val)}
		case bool:
			av[k] = &types.AttributeValueMemberBOOL{Value: val}
		case map[string]interface{}:
			av[k] = &types.AttributeValueMemberM{Value: interfaceMapToAttributeValue(val)}
		case []interface{}:
			items := make([]types.AttributeValue, 0, len(val))
			for _, itm := range val {
				if m, ok := itm.(map[string]interface{}); ok {
					items = append(items, &types.AttributeValueMemberM{Value: interfaceMapToAttributeValue(m)})
				} else if s, ok := itm.(string); ok {
					items = append(items, &types.AttributeValueMemberS{Value: s})
				}
			}
			av[k] = &types.AttributeValueMemberL{Value: items}
		default:
			// Unsupported type, skip
		}
	}
	return av
}

// NewDynamoDBClient creates a DynamoDBClient using IRSA / DefaultCredentialChain.
// No static credentials are accepted — the SDK resolves credentials from the
// pod's service-account token via the AWS_ROLE_ARN + AWS_WEB_IDENTITY_TOKEN_FILE
// env vars injected by the EKS IRSA admission webhook.
func NewDynamoDBClient(ctx context.Context, region, tableName string, logger zerolog.Logger) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	logger.Info().Str("region", region).Str("table", tableName).Msg("AWS DynamoDB client initialized")
	return &DynamoDBClient{
		api:       dynamodb.NewFromConfig(cfg),
		tableName: tableName,
		logger:    logger,
	}, nil
}

// NewDynamoDBClientFromAPI creates a DynamoDBClient backed by the provided DynamoDBAPI.
// Used in unit tests to inject a mock without real AWS credentials.
func NewDynamoDBClientFromAPI(api DynamoDBAPI, tableName string, logger zerolog.Logger) *DynamoDBClient {
	return &DynamoDBClient{
		api:       api,
		tableName: tableName,
		logger:    logger,
	}
}

// Get retrieves a single item by its ID primary key.
// Returns ErrItemNotFound when the item does not exist in the table.
func (d *DynamoDBClient) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	d.logger.Debug().Str("id", id).Msg("DynamoDB GetItem")
	result, err := d.api.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &d.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		d.logger.Error().Err(err).Str("id", id).Msg("DynamoDB GetItem failed")
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	if result.Item == nil {
		d.logger.Warn().Str("id", id).Msg("DynamoDB item not found")
		return nil, ErrItemNotFound
	}
	return avMapToInterface(result.Item), nil
}

// Create persists a new item via PutItem.
// The caller is responsible for marshaling the domain entity to AttributeValue map.
func (d *DynamoDBClient) Create(ctx context.Context, item map[string]interface{}) error {
	d.logger.Debug().Interface("item", item).Msg("DynamoDB PutItem")
	avItem := interfaceMapToAttributeValue(item)
	_, err := d.api.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &d.tableName,
		Item:      avItem,
	})
	if err != nil {
		d.logger.Error().Err(err).Msg("DynamoDB PutItem failed")
		return fmt.Errorf("failed to put item: %w", err)
	}
	return nil
}

// Query returns all items in the table via a full Scan.
// NOTE: This performs a full table Scan with no LIMIT or filter expression.
// For production use, add a Limit parameter or filter expression to avoid excessive read capacity consumption.
// For production use consider adding filter expressions or paginated queries.
func (d *DynamoDBClient) Query(ctx context.Context) ([]map[string]interface{}, error) {
	d.logger.Debug().Str("table", d.tableName).Msg("DynamoDB Scan")
	result, err := d.api.Scan(ctx, &dynamodb.ScanInput{
		TableName: &d.tableName,
	})
	if err != nil {
		d.logger.Error().Err(err).Msg("DynamoDB Scan failed")
		return nil, fmt.Errorf("failed to scan items: %w", err)
	}
	var items []map[string]interface{}
	for _, av := range result.Items {
		items = append(items, avMapToInterface(av))
	}
	return items, nil
}

// Delete removes an item by its ID primary key.
func (d *DynamoDBClient) Delete(ctx context.Context, id string) error {
	d.logger.Debug().Str("id", id).Msg("DynamoDB DeleteItem")
	_, err := d.api.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &d.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		d.logger.Error().Err(err).Str("id", id).Msg("DynamoDB DeleteItem failed")
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

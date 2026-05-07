package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

type DynamoDBClient struct {
	client    *dynamodb.Client
	tableName string
	logger    zerolog.Logger
}

func NewDynamoDBClient(ctx context.Context, region, tableName string, logger zerolog.Logger) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	logger.Info().Str("region", region).Str("table", tableName).Msg("AWS DynamoDB client initialized")
	return &DynamoDBClient{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
		logger:    logger,
	}, nil
}

func (d *DynamoDBClient) Get(ctx context.Context, id string) (map[string]types.AttributeValue, error) {
	d.logger.Debug().Str("id", id).Msg("DynamoDB GetItem")
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
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
		return nil, fmt.Errorf("item not found")
	}
	return result.Item, nil
}

func (d *DynamoDBClient) Create(ctx context.Context, item map[string]types.AttributeValue) error {
	d.logger.Debug().Interface("item", item).Msg("DynamoDB PutItem")
	_, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &d.tableName,
		Item:      item,
	})
	if err != nil {
		d.logger.Error().Err(err).Msg("DynamoDB PutItem failed")
		return fmt.Errorf("failed to put item: %w", err)
	}
	return nil
}

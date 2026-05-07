package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/rs/zerolog"
)

type CosmosDBClient struct {
	container *azcosmos.ContainerClient
	logger    zerolog.Logger
}

func NewCosmosDBClient(ctx context.Context, endpoint, database, container string, logger zerolog.Logger) (*CosmosDBClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure credential: %w", err)
	}

	client, err := azcosmos.NewClient(endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create CosmosDB client: %w", err)
	}

	db, err := client.NewDatabase(database)
	if err != nil {
		return nil, err
	}

	cont, err := db.NewContainer(container)
	if err != nil {
		return nil, err
	}

	logger.Info().Str("endpoint", endpoint).Str("database", database).Str("container", container).Msg("Azure CosmosDB client initialized")
	return &CosmosDBClient{container: cont, logger: logger}, nil
}

func (c *CosmosDBClient) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	partitionKey := "task"
	c.logger.Debug().Str("id", id).Msg("CosmosDB ReadItem")
	response, err := c.container.ReadItem(ctx, azcosmos.NewPartitionKeyString(partitionKey), id, nil)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("CosmosDB ReadItem failed")
		return nil, fmt.Errorf("failed to read item: %w", err)
	}

	var item map[string]interface{}
	if err := response.Unmarshal(&item); err != nil {
		return nil, err
	}
	return item, nil
}

func (c *CosmosDBClient) Create(ctx context.Context, id string, item interface{}) error {
	partitionKey := "task"
	c.logger.Debug().Str("id", id).Msg("CosmosDB CreateItem")
	_, err := c.container.CreateItem(ctx, azcosmos.NewPartitionKeyString(partitionKey), item, nil)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("CosmosDB CreateItem failed")
		return fmt.Errorf("failed to create item: %w", err)
	}
	return nil
}

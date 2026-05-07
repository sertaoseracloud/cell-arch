// Package azure provides the Azure CosmosDB adapter for the sidecar.
// It authenticates via Workload Identity (no static secrets).
package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/rs/zerolog"
)

// CosmosDBClient wraps an azcosmos ContainerClient for task persistence.
type CosmosDBClient struct {
	container *azcosmos.ContainerClient
	logger    zerolog.Logger
}

// NewCosmosDBClient constructs a CosmosDBClient.
// Authentication is via DefaultAzureCredential (Workload Identity / IRSA equivalent).
func NewCosmosDBClient(ctx context.Context, endpoint, database, container string, logger zerolog.Logger) (*CosmosDBClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure credential: %w", err)
	}

	client, err := azcosmos.NewClient(endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create CosmosDB client: %w", err)
	}

	cont, err := client.NewContainer(database, container)
	if err != nil {
		return nil, fmt.Errorf("failed to get CosmosDB container: %w", err)
	}

	logger.Info().
		Str("endpoint", endpoint).
		Str("database", database).
		Str("container", container).
		Msg("Azure CosmosDB client initialized")
	return &CosmosDBClient{container: cont, logger: logger}, nil
}

// Get retrieves an item by ID from CosmosDB.
func (c *CosmosDBClient) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	const partitionKey = "task"
	c.logger.Debug().Str("id", id).Msg("CosmosDB ReadItem")
	response, err := c.container.ReadItem(ctx, azcosmos.NewPartitionKeyString(partitionKey), id, nil)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("CosmosDB ReadItem failed")
		return nil, fmt.Errorf("failed to read item: %w", err)
	}

	var item map[string]interface{}
	if err := json.Unmarshal(response.Value, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CosmosDB response: %w", err)
	}
	return item, nil
}

// Create persists an item in CosmosDB.
// The item must be JSON-serialisable.
func (c *CosmosDBClient) Create(ctx context.Context, id string, item interface{}) error {
	const partitionKey = "task"
	c.logger.Debug().Str("id", id).Msg("CosmosDB CreateItem")

	payload, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item for CosmosDB: %w", err)
	}

	_, err = c.container.CreateItem(ctx, azcosmos.NewPartitionKeyString(partitionKey), payload, nil)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("CosmosDB CreateItem failed")
		return fmt.Errorf("failed to create item: %w", err)
	}
	return nil
}

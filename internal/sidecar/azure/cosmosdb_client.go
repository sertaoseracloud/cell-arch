// Package azure provides the Azure CosmosDB adapter for the sidecar.
// It authenticates via Workload Identity (no static secrets).
package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/rs/zerolog"
)

// CosmosDBClient wraps an azcosmos ContainerClient for task persistence.

// ErrItemNotFound is returned when a requested item does not exist in CosmosDB.
var ErrItemNotFound = errors.New("item not found")

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

// isNotFoundErr checks if the error message indicates a not-found condition.
// It looks for common Azure CosmosDB error phrases. This is a best-effort
// detection because the SDK version in use does not expose typed error codes.
func isNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "404") ||
		strings.Contains(msg, "Not Found") ||
		strings.Contains(msg, "ResourceNotFound")
}

// Get retrieves an item by ID from CosmosDB.
func (c *CosmosDBClient) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	const partitionKey = "task"
	c.logger.Debug().Str("id", id).Msg("CosmosDB ReadItem")
	response, err := c.container.ReadItem(ctx, azcosmos.NewPartitionKeyString(partitionKey), id, nil)
	if err != nil {
		if isNotFoundErr(err) {
			c.logger.Warn().Str("id", id).Msg("CosmosDB item not found")
			return nil, ErrItemNotFound
		}
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
// The item must be JSON-serialisable and contain an "id" field.
func (c *CosmosDBClient) Create(ctx context.Context, item map[string]interface{}) error {
	const partitionKey = "task"
	// Extract ID from item map
	idVal, ok := item["id"]
	if !ok {
		return fmt.Errorf("item missing id field")
	}
	idStr := fmt.Sprintf("%v", idVal)
	c.logger.Debug().Str("id", idStr).Msg("CosmosDB CreateItem")

	payload, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item for CosmosDB: %w", err)
	}

	_, err = c.container.CreateItem(ctx, azcosmos.NewPartitionKeyString(partitionKey), payload, nil)
	if err != nil {
		c.logger.Error().Err(err).Str("id", idStr).Msg("CosmosDB CreateItem failed")
		return fmt.Errorf("failed to create item: %w", err)
	}
	return nil
}

// Query returns all items from the CosmosDB container (no filter).
// NOTE: This performs a full container scan with no LIMIT or filter.
// For production use, add filtering and pagination to avoid performance issues on large datasets.
func (c *CosmosDBClient) Query(ctx context.Context) ([]map[string]interface{}, error) {
	c.logger.Debug().Msg("CosmosDB QueryItems")
	query := "SELECT * FROM c"
	pager := c.container.NewQueryItemsPager(query, azcosmos.NewPartitionKey(), nil)

	var items []map[string]interface{}
	for pager.More() {
		response, err := pager.NextPage(ctx)
		if err != nil {
			c.logger.Error().Err(err).Msg("CosmosDB QueryItems failed")
			return nil, fmt.Errorf("failed to query items: %w", err)
		}
		for _, val := range response.Items {
			var item map[string]interface{}
			if err := json.Unmarshal(val, &item); err != nil {
				return nil, fmt.Errorf("failed to unmarshal query result: %w", err)
			}
			items = append(items, item)
		}
	}
	return items, nil
}

// Delete removes an item by ID from CosmosDB.
func (c *CosmosDBClient) Delete(ctx context.Context, id string) error {
	const partitionKey = "task"
	c.logger.Debug().Str("id", id).Msg("CosmosDB DeleteItem")
	_, err := c.container.DeleteItem(ctx, azcosmos.NewPartitionKeyString(partitionKey), id, nil)
	if err != nil {
		// Check if not found
		if isNotFoundErr(err) {
			c.logger.Warn().Str("id", id).Msg("CosmosDB item not found on delete")
			return ErrItemNotFound
		}
		c.logger.Error().Err(err).Str("id", id).Msg("CosmosDB DeleteItem failed")
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

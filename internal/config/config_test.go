package config_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/cell-arch/internal/config"
)

func TestLoadAppConfig_Defaults(t *testing.T) {
	t.Setenv("GRPC_PORT", "")
	t.Setenv("SIDECAR_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("APP_ENV", "")

	cfg, err := config.LoadAppConfig(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "localhost:50051", cfg.SidecarAddr)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, 8080, cfg.GRPCPort)
}

func TestLoadAppConfig_FromEnv(t *testing.T) {
	t.Setenv("GRPC_PORT", "9090")
	t.Setenv("SIDECAR_ADDR", "localhost:50052")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("APP_ENV", "production")

	cfg, err := config.LoadAppConfig(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "localhost:50052", cfg.SidecarAddr)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, 9090, cfg.GRPCPort)
}

func TestLoadAppConfig_InvalidPort(t *testing.T) {
	t.Setenv("GRPC_PORT", "not-a-port")

	_, err := config.LoadAppConfig(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GRPC_PORT")
}

func TestLoadSidecarConfig_Defaults(t *testing.T) {
	t.Setenv("GRPC_PORT", "")
	t.Setenv("TLS_CERT", "")
	t.Setenv("TLS_KEY", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("DYNAMODB_TABLE", "")
	t.Setenv("AZURE_COSMOS_ENDPOINT", "")
	t.Setenv("COSMOS_DATABASE", "")
	t.Setenv("COSMOS_CONTAINER", "")

	cfg, err := config.LoadSidecarConfig(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 50051, cfg.GRPCPort)
	assert.Equal(t, "us-east-1", cfg.AWSRegion)
	assert.Equal(t, "tasks", cfg.DynamoDBTable)
	assert.Equal(t, "tasks", cfg.CosmosDatabase)
	assert.Equal(t, "tasks", cfg.CosmosContainer)
	assert.Empty(t, cfg.TLSCert)
	assert.Empty(t, cfg.TLSKey)
}

func TestLoadSidecarConfig_FromEnv(t *testing.T) {
	t.Setenv("GRPC_PORT", "50055")
	t.Setenv("TLS_CERT", "/certs/tls.crt")
	t.Setenv("TLS_KEY", "/certs/tls.key")
	t.Setenv("AWS_REGION", "eu-west-1")
	t.Setenv("DYNAMODB_TABLE", "my-table")
	t.Setenv("AZURE_COSMOS_ENDPOINT", "https://myaccount.documents.azure.com")
	t.Setenv("COSMOS_DATABASE", "mydb")
	t.Setenv("COSMOS_CONTAINER", "mycontainer")

	cfg, err := config.LoadSidecarConfig(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 50055, cfg.GRPCPort)
	assert.Equal(t, "/certs/tls.crt", cfg.TLSCert)
	assert.Equal(t, "/certs/tls.key", cfg.TLSKey)
	assert.Equal(t, "eu-west-1", cfg.AWSRegion)
	assert.Equal(t, "my-table", cfg.DynamoDBTable)
	assert.Equal(t, "https://myaccount.documents.azure.com", cfg.AzureCosmosEndpoint)
	assert.Equal(t, "mydb", cfg.CosmosDatabase)
	assert.Equal(t, "mycontainer", cfg.CosmosContainer)
}

func TestLoadSidecarConfig_InvalidPort(t *testing.T) {
	t.Setenv("GRPC_PORT", "bad-port")

	_, err := config.LoadSidecarConfig(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GRPC_PORT")
}

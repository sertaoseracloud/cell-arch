package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/cell-arch/internal/sidecar/config"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Point to a non-existent file so defaults apply.
	t.Setenv("SIDECAR_CONFIG", filepath.Join(t.TempDir(), "nonexistent.yaml"))

	cfg, err := config.LoadConfig(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 50051, cfg.GRPCPort)
	assert.Equal(t, "us-east-1", cfg.Region)
	assert.Equal(t, "tasks", cfg.DynamoDBTable)
	assert.Equal(t, "tasks", cfg.CosmosDatabase)
	assert.Equal(t, "tasks", cfg.CosmosContainer)
	assert.Empty(t, cfg.TLSCert)
	assert.Empty(t, cfg.TLSKey)
}

func TestLoadConfig_FromYAML(t *testing.T) {
	yaml := `
grpc_port: 9090
tls_cert: "/certs/server.crt"
tls_key: "/certs/server.key"
tls_ca_cert: "/certs/ca.crt"
aws_endpoint: "http://localhost:4566"
azure_endpoint: "https://account.documents.azure.com:443/"
region: "eu-west-1"
dynamodb_table: "prod-tasks"
cosmos_database: "proddb"
cosmos_container: "prodcont"
`
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "sidecar.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(yaml), 0o600))
	t.Setenv("SIDECAR_CONFIG", cfgPath)

	cfg, err := config.LoadConfig(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 9090, cfg.GRPCPort)
	assert.Equal(t, "/certs/server.crt", cfg.TLSCert)
	assert.Equal(t, "/certs/server.key", cfg.TLSKey)
	assert.Equal(t, "/certs/ca.crt", cfg.TLSCACert)
	assert.Equal(t, "http://localhost:4566", cfg.AWSEndpoint)
	assert.Equal(t, "https://account.documents.azure.com:443/", cfg.AzureEndpoint)
	assert.Equal(t, "eu-west-1", cfg.Region)
	assert.Equal(t, "prod-tasks", cfg.DynamoDBTable)
	assert.Equal(t, "proddb", cfg.CosmosDatabase)
	assert.Equal(t, "prodcont", cfg.CosmosContainer)
}

func TestLoadConfig_EnvOverridesYAML(t *testing.T) {
	yaml := `grpc_port: 9090
region: "eu-west-1"
dynamodb_table: "file-table"
`
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "sidecar.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(yaml), 0o600))
	t.Setenv("SIDECAR_CONFIG", cfgPath)

	// Env vars override the file.
	t.Setenv("GRPC_PORT", "7777")
	t.Setenv("AWS_REGION", "ap-southeast-1")
	t.Setenv("DYNAMODB_TABLE", "env-table")
	t.Setenv("TLS_CERT", "/env/cert.crt")
	t.Setenv("TLS_KEY", "/env/cert.key")

	cfg, err := config.LoadConfig(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 7777, cfg.GRPCPort)
	assert.Equal(t, "ap-southeast-1", cfg.Region)
	assert.Equal(t, "env-table", cfg.DynamoDBTable)
	assert.Equal(t, "/env/cert.crt", cfg.TLSCert)
	assert.Equal(t, "/env/cert.key", cfg.TLSKey)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "sidecar.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte("grpc_port: [invalid"), 0o600))
	t.Setenv("SIDECAR_CONFIG", cfgPath)

	_, err := config.LoadConfig(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestLoadConfig_ReadError(t *testing.T) {
	// Simulate a read error by pointing SIDECAR_CONFIG at a directory,
	// which os.ReadFile cannot read as a file (works cross-platform).
	dir := t.TempDir()
	t.Setenv("SIDECAR_CONFIG", dir) // dir, not a file

	_, err := config.LoadConfig(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read")
}

func TestLoadConfig_AzureAndCosmosEnvOverrides(t *testing.T) {
	t.Setenv("SIDECAR_CONFIG", filepath.Join(t.TempDir(), "nope.yaml"))
	t.Setenv("AZURE_COSMOS_ENDPOINT", "https://myaccount.documents.azure.com:443/")
	t.Setenv("COSMOS_DATABASE", "mydb")
	t.Setenv("COSMOS_CONTAINER", "mycont")
	t.Setenv("TLS_CA_CERT", "/ca.crt")
	t.Setenv("AWS_ENDPOINT", "http://localhost:4566")

	cfg, err := config.LoadConfig(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "https://myaccount.documents.azure.com:443/", cfg.AzureEndpoint)
	assert.Equal(t, "mydb", cfg.CosmosDatabase)
	assert.Equal(t, "mycont", cfg.CosmosContainer)
	assert.Equal(t, "/ca.crt", cfg.TLSCACert)
	assert.Equal(t, "http://localhost:4566", cfg.AWSEndpoint)
}

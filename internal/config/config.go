// Package config provides typed application configuration loaded from
// environment variables. No init() side effects — call LoadConfig explicitly.
package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
)

// AppConfig holds the main application configuration.
type AppConfig struct {
	// SidecarAddr is the address of the local sidecar gRPC server.
	SidecarAddr string
	// LogLevel controls the verbosity of the structured logger.
	LogLevel string
	// Env identifies the runtime environment (development, production, etc.)
	Env string
	// GRPCPort is the port the main app listens on for its own gRPC server.
	GRPCPort int
}

// SidecarConfig holds the sidecar proxy configuration.
type SidecarConfig struct {
	// GRPCPort is the port the sidecar gRPC server listens on.
	GRPCPort int
	// TLSCert is the path to the TLS certificate file.
	TLSCert string
	// TLSKey is the path to the TLS key file.
	TLSKey string
	// AWSRegion is the AWS region for DynamoDB.
	AWSRegion string
	// DynamoDBTable is the DynamoDB table name.
	DynamoDBTable string
	// AzureCosmosEndpoint is the CosmosDB account endpoint URL.
	AzureCosmosEndpoint string
	// CosmosDatabase is the CosmosDB database name.
	CosmosDatabase string
	// CosmosContainer is the CosmosDB container name.
	CosmosContainer string
}

// LoadAppConfig reads app configuration from environment variables.
// ctx is accepted at the boundary per D-03 and may carry a deadline for
// slow external config sources in future.
func LoadAppConfig(_ context.Context) (*AppConfig, error) {
	port, err := parseInt(getEnvOrDefault("GRPC_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid GRPC_PORT: %w", err)
	}
	return &AppConfig{
		SidecarAddr: getEnvOrDefault("SIDECAR_ADDR", "localhost:50051"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
		Env:         getEnvOrDefault("APP_ENV", "development"),
		GRPCPort:    port,
	}, nil
}

// LoadSidecarConfig reads sidecar configuration from environment variables.
func LoadSidecarConfig(_ context.Context) (*SidecarConfig, error) {
	port, err := parseInt(getEnvOrDefault("GRPC_PORT", "50051"))
	if err != nil {
		return nil, fmt.Errorf("invalid GRPC_PORT: %w", err)
	}
	return &SidecarConfig{
		GRPCPort:            port,
		TLSCert:             os.Getenv("TLS_CERT"),
		TLSKey:              os.Getenv("TLS_KEY"),
		AWSRegion:           getEnvOrDefault("AWS_REGION", "us-east-1"),
		DynamoDBTable:       getEnvOrDefault("DYNAMODB_TABLE", "tasks"),
		AzureCosmosEndpoint: os.Getenv("AZURE_COSMOS_ENDPOINT"),
		CosmosDatabase:      getEnvOrDefault("COSMOS_DATABASE", "tasks"),
		CosmosContainer:     getEnvOrDefault("COSMOS_CONTAINER", "tasks"),
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func parseInt(s string) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as integer: %w", s, err)
	}
	return v, nil
}

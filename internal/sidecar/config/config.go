// Package config provides sidecar-specific configuration loaded from sidecar.yaml.
// Environment variables take precedence over file values, following the
// twelve-factor app principle. Call LoadConfig explicitly — no init() side effects (D-02).
package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "sidecar.yaml"

// Config holds the full sidecar configuration.
// Fields map 1-to-1 to sidecar.yaml keys and environment variable overrides.
type Config struct {
	// GRPCPort is the port the sidecar gRPC server listens on (default 50051).
	GRPCPort int `yaml:"grpc_port"`
	// TLSCert is the path to the server TLS certificate (PEM). Required for mTLS (D-15).
	TLSCert string `yaml:"tls_cert"`
	// TLSKey is the path to the server TLS private key (PEM). Required for mTLS (D-15).
	TLSKey string `yaml:"tls_key"`
	// TLSCACert is the path to the CA certificate used to verify client certs (optional).
	TLSCACert string `yaml:"tls_ca_cert"`
	// AWSEndpoint is an optional override for the DynamoDB endpoint (e.g. LocalStack).
	AWSEndpoint string `yaml:"aws_endpoint"`
	// AzureEndpoint is the CosmosDB account endpoint URL.
	AzureEndpoint string `yaml:"azure_endpoint"`
	// Region is the AWS region for DynamoDB.
	Region string `yaml:"region"`
	// DynamoDBTable is the DynamoDB table name.
	DynamoDBTable string `yaml:"dynamodb_table"`
	// CosmosDatabase is the CosmosDB database name.
	CosmosDatabase string `yaml:"cosmos_database"`
	// CosmosContainer is the CosmosDB container name.
	CosmosContainer string `yaml:"cosmos_container"`
}

// LoadConfig reads sidecar configuration from sidecar.yaml and applies
// environment-variable overrides. ctx is accepted per D-03.
//
// File path can be overridden with the SIDECAR_CONFIG env var.
// If the file does not exist, defaults are used (env vars still apply).
func LoadConfig(_ context.Context) (*Config, error) {
	cfg := defaults()

	path := os.Getenv("SIDECAR_CONFIG")
	if path == "" {
		path = defaultConfigPath
	}

	if data, err := os.ReadFile(path); err == nil {
		// File exists — parse YAML into cfg.
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		// File exists but cannot be read — surface the error.
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	// If file does not exist, cfg keeps its defaults.

	// Environment variables override file values (twelve-factor app).
	applyEnvOverrides(cfg)

	return cfg, nil
}

// defaults returns a Config with sensible defaults for local development.
func defaults() *Config {
	return &Config{
		GRPCPort:        50051,
		Region:          "us-east-1",
		DynamoDBTable:   "tasks",
		CosmosDatabase:  "tasks",
		CosmosContainer: "tasks",
	}
}

// applyEnvOverrides overwrites cfg fields with env var values when set.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("GRPC_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.GRPCPort = port
		}
	}
	if v := os.Getenv("TLS_CERT"); v != "" {
		cfg.TLSCert = v
	}
	if v := os.Getenv("TLS_KEY"); v != "" {
		cfg.TLSKey = v
	}
	if v := os.Getenv("TLS_CA_CERT"); v != "" {
		cfg.TLSCACert = v
	}
	if v := os.Getenv("AWS_ENDPOINT"); v != "" {
		cfg.AWSEndpoint = v
	}
	if v := os.Getenv("AZURE_COSMOS_ENDPOINT"); v != "" {
		cfg.AzureEndpoint = v
	}
	if v := os.Getenv("AWS_REGION"); v != "" {
		cfg.Region = v
	}
	if v := os.Getenv("DYNAMODB_TABLE"); v != "" {
		cfg.DynamoDBTable = v
	}
	if v := os.Getenv("COSMOS_DATABASE"); v != "" {
		cfg.CosmosDatabase = v
	}
	if v := os.Getenv("COSMOS_CONTAINER"); v != "" {
		cfg.CosmosContainer = v
	}
}

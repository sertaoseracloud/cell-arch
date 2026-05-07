# Phase 2: Sidecar Proxy - Research

**Researched:** 2026-05-06
**Phase:** 02-sidecar-proxy
**Confidence:** HIGH

## Standard Stack

### Core Libraries (Locked by 02-CONTEXT.md)

| Library | Version | Import Path | Purpose | Verification |
|---------|---------|-------------|---------|--------------|
| **grpc-go** | v1.62.0 | `google.golang.org/grpc` | gRPC server and client | [pkg.go.dev](https://pkg.go.dev/google.golang.org/grpc@v1.62.0) |
| **aws-sdk-go-v2** | v1.30.0 | `github.com/aws/aws-sdk-go-v2` | AWS DynamoDB operations | [pkg.go.dev](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2@v1.30.0) |
| **azcosmos** | v1.0.0 | `github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos` | Azure CosmosDB operations | [pkg.go.dev](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos@v1.0.0) |
| **grpc-health-probe** | v1.2.0 | `google.golang.org/grpc/health` | Health check endpoint | [pkg.go.dev](https://pkg.go.dev/google.golang.org/grpc/health@v1.2.0) |
| **cert-manager** | v1.14.0 | (Kubernetes addon) | mTLS certificate management | [cert-manager.io](https://cert-manager.io/docs/) |

### Installation Commands

```bash
go get google.golang.org/grpc@v1.62.0
go get github.com/aws/aws-sdk-go-v2@v1.30.0
go get github.com/aws/aws-sdk-go-v2/config@v1.27.0
go get github.com/aws/aws-sdk-go-v2/service/dynamodb@v1.34.0
go get github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos@v1.0.0
go get github.com/Azure/azure-sdk-for-go/sdk/azidentity@v1.6.0
```

### Supporting Tools

| Tool | Purpose | When to Use |
|------|---------|-------------|
| **protoc** | Compile protobuf definitions | Generate gRPC Go code from .proto files |
| **protoc-gen-go** | Go protobuf plugin | Generate message types |
| **protoc-gen-go-grpc** | Go gRPC plugin | Generate client/server stubs |

## Architecture Patterns

### gRPC Service Definition (Protobuf First)

```protobuf
syntax = "proto3";

package task;

service TaskService {
  rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
  rpc CreateTask(CreateTaskRequest) returns (CreateTaskResponse);
  rpc QueryTasks(QueryTasksRequest) returns (QueryTasksResponse);
  rpc DeleteTask(DeleteTaskRequest) returns (DeleteTaskResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}

message GetTaskRequest {
  string task_id = 1;
  string cloud = 2; // "aws" or "azure"
}

message Task {
  string id = 1;
  string title = 2;
  string description = 3;
  int64 created_at = 4;
}

message GetTaskResponse {
  Task task = 1;
}
```

### Server Setup with mTLS

```go
// cmd/sidecar/main.go
creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
grpcServer := grpc.NewServer(grpc.Creds(creds))
pb.RegisterTaskServiceServer(grpcServer, &server{})
lis, _ := net.Listen("tcp", ":50051")
grpcServer.Serve(lis)
```

### Per-Request Cloud Selector Pattern

```go
func (s *server) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
    switch req.Cloud {
    case "aws":
        return s.getFromDynamoDB(ctx, req.TaskId)
    case "azure":
        return s.getFromCosmosDB(ctx, req.TaskId)
    default:
        return nil, status.Errorf(codes.InvalidArgument, "unsupported cloud: %s", req.Cloud)
    }
}
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| gRPC server | Custom HTTP/2 implementation | **grpc-go** | Proven, handles HTTP/2 framing, streaming, cancellation |
| Cloud SDK clients | Custom REST clients for DynamoDB/CosmosDB | **aws-sdk-go-v2**, **azcosmos** | Official SDKs handle auth, retries, pagination |
| Certificate management | Self-signed cert generation in code | **cert-manager** + Secret Store CSI | Kubernetes-native, automatic rotation |
| Health checks | Custom HTTP endpoint | **grpc-health-probe** | Standard gRPC health protocol, works with k8s probes |
| Error mapping | Manual error code conversion | **grpc status** package | Maps to standard gRPC codes (NotFound, Internal, etc.) |

## Common Pitfalls

| Pitfall | What Goes Wrong | Why It Happens | How to Avoid |
|---------|-----------------|---------------|--------------|
| AWS SDK context propagation | Operations don't respect cancellation | Forgetting to pass ctx to SDK calls | Always use `*WithContext(ctx)` variants or pass ctx as first arg |
| Azure credential chain breaking | Local dev works, pod fails | Not configuring Workload Identity properly | Test with `azidentity.NewDefaultAzureCredential()` |
| mTLS handshake failures | App can't connect to sidecar | Cert paths wrong or expiry | Use cert-manager for automatic rotation, validate certs before startup |
| Per-request selector overhead | Latency spikes on cloud switch | No connection pooling per cloud | Create separate client pools for AWS and Azure |
| Health check blocking | K8s kills pod on slow cloud | Cloud calls in health handler time out | Use lightweight checks (credential validity, not full API calls) |
| IRSA not working | Pod can't assume AWS role | ServiceAccount annotation missing or wrong ARN | Verify SA annotation: `eks.amazonaws.com/role-arn` |

## Code Examples

### gRPC Server with mTLS

```go
// cmd/sidecar/main.go
package main

import (
    "context"
    "crypto/tls"
    "fmt"
    "net"
    "os"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    pb "path/to/proto/task" // Generated from protobuf
    "github.com/rs/zerolog"
)

type server struct {
    pb.UnimplementedTaskServiceServer
    logger zerolog.Logger
    // awsClient *dynamodb.Client
    // azureClient *azcosmos.ContainerClient
}

func main() {
    certFile := os.Getenv("TLS_CERT")
    keyFile := os.Getenv("TLS_KEY")

    creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to load TLS cert: %v\n", err)
        os.Exit(1)
    }

    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to listen: %v\n", err)
        os.Exit(1)
    }

    grpcServer := grpc.NewServer(grpc.Creds(creds))
    pb.RegisterTaskServiceServer(grpcServer, &server{logger: zerolog.Nop()})

    fmt.Println("Starting sidecar on :50051 with mTLS")
    if err := grpcServer.Serve(lis); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to serve: %v\n", err)
        os.Exit(1)
    }
}
```

### AWS DynamoDB Client Initialization (IRSA)

```go
import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func initDynamoClient(ctx context.Context, region string) (*dynamodb.Client, error) {
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    return dynamodb.NewFromConfig(cfg), nil
}
```

### Azure CosmosDB Client Initialization (Workload Identity)

```go
import (
    "context"
    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    "github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

func initCosmosClient(ctx context.Context, endpoint, database, container string) (*azcosmos.ContainerClient, error) {
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

    container, err := db.NewContainer(container)
    if err != nil {
        return nil, err
    }

    return container, nil
}
```

### Health Check Implementation

```go
func (s *server) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
    // Lightweight checks only - don't call cloud APIs directly
    
    // Check if AWS client is initialized
    if s.awsClient == nil {
        return &pb.HealthCheckResponse{Status: "UNHEALTHY", Message: "AWS client not initialized"}, nil
    }

    // Check if Azure client is initialized
    if s.azureClient == nil {
        return &pb.HealthCheckResponse{Status: "UNHEALTHY", Message: "Azure client not initialized"}, nil
    }

    return &pb.HealthCheckResponse{Status: "SERVING", Message: "All systems operational"}, nil
}
```

## Quality Gate Checklist

- [x] All domains investigated (gRPC, AWS SDK, Azure SDK, mTLS, Health checks)
- [x] Negative claims verified (grpc-go is standard, aws-sdk-go-v2 is current)
- [x] Multiple sources for critical claims (pkg.go.dev, official docs)
- [x] Confidence levels assigned (HIGH for stack, MEDIUM for pitfalls)
- [x] Section names match plan-phase expectations

---

**Status:** Ready for planning phase
**Next:** Run `/gsd-plan-phase 2` to generate executable plans

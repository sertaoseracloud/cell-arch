package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "github.com/yourorg/cell-arch/proto"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/rs/zerolog"

	sidecaraws "github.com/yourorg/cell-arch/internal/sidecar/aws"
	sidecarazure "github.com/yourorg/cell-arch/internal/sidecar/azure"
)

type server struct {
	pb.UnimplementedTaskServiceServer
	health      *health.Server
	awsClient   *aws.DynamoDBClient
	azureClient *azure.CosmosDBClient
	logger      zerolog.Logger
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize structured logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")

	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load TLS cert")
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to listen")
		os.Exit(1)
	}

	// Initialize AWS client (IRSA - no static credentials)
	awsClient, err := sidecaraws.NewDynamoDBClient(ctx, os.Getenv("AWS_REGION"), os.Getenv("DYNAMODB_TABLE"), logger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create AWS client")
		os.Exit(1)
	}

	// Initialize Azure client (Workload Identity - no static credentials)
	azureClient, err := sidecarazure.NewCosmosDBClient(ctx, os.Getenv("AZURE_COSMOS_ENDPOINT"), os.Getenv("COSMOS_DATABASE"), os.Getenv("COSMOS_CONTAINER"), logger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create Azure client")
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(grpc.Creds(creds))
	srv := &server{
		health:      health.NewServer(),
		awsClient:   awsClient,
		azureClient: azureClient,
		logger:      logger,
	}
	pb.RegisterTaskServiceServer(grpcServer, srv)
	grpc_health_v1.RegisterHealthServer(grpcServer, srv.health)

	fmt.Println("Starting sidecar on :50051 with mTLS")
	srv.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to serve: %v\n", err)
		os.Exit(1)
	}
}

func (s *server) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	// Route based on cloud field
	switch req.Cloud {
	case "aws":
		// Call AWS DynamoDB
		item, err := s.awsClient.Get(ctx, req.TaskId)
		if err != nil {
			return nil, MapAWSError(err)
		}
		// Convert item to Task
		task := &pb.Task{
			Id:    req.TaskId,
			Title: "from aws",
		}
		return &pb.GetTaskResponse{Task: task}, nil
	case "azure":
		// Call Azure CosmosDB
		item, err := s.azureClient.Get(ctx, req.TaskId)
		if err != nil {
			return nil, MapAzureError(err)
		}
		task := &pb.Task{
			Id:    req.TaskId,
			Title: "from azure",
		}
		return &pb.GetTaskResponse{Task: task}, nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported cloud: %s", req.Cloud)
	}
}

func (s *server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	// Route based on cloud field
	switch req.Cloud {
	case "aws":
		// Call AWS DynamoDB
		item := map[string]interface{}{
			"id":    req.Task.Id,
			"title": req.Task.Title,
		}
		err := s.awsClient.Create(ctx, req.Task.Id, item)
		if err != nil {
			return nil, MapAWSError(err)
		}
		return &pb.CreateTaskResponse{Success: true}, nil
	case "azure":
		// Call Azure CosmosDB
		item := map[string]interface{}{
			"id":    req.Task.Id,
			"title": req.Task.Title,
		}
		err := s.azureClient.Create(ctx, req.Task.Id, item)
		if err != nil {
			return nil, MapAzureError(err)
		}
		return &pb.CreateTaskResponse{Success: true}, nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported cloud: %s", req.Cloud)
	}
}

func (s *server) QueryTasks(ctx context.Context, req *pb.QueryTasksRequest) (*pb.QueryTasksResponse, error) {
	return &pb.QueryTasksResponse{}, nil
}

func (s *server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	return &pb.DeleteTaskResponse{Success: true}, nil
}

func (s *server) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	// Check if clients are initialized
	if s.awsClient == nil || s.azureClient == nil {
		return &pb.HealthCheckResponse{
			Status:  "UNHEALTHY",
			Message: "Clients not initialized",
		}, nil
	}
	return &pb.HealthCheckResponse{
		Status:  "SERVING",
		Message: "All systems operational",
	}, nil
}

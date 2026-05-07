---
phase: 2
plan: 3
subsystem: sidecar-proxy
tags: [go, azure, cosmosdb, sidecar, cloud-backend]
dependency_graph:
  requires: [phase-2-plan-01-grpc-server, phase-2-plan-02-dynamodb]
  provides: [cosmosdb-client, azure-backend]
  affects: [phase-2-plan-04-cloud-selector, phase-3-terraform]
tech_stack:
  added:
    - github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos
    - github.com/Azure/azure-sdk-for-go/sdk/azidentity
  patterns:
    - Workload Identity via DefaultAzureCredential (no static secrets, D-10)
    - Generic CloudBackend interface with map[string]interface{} (D-18)
    - Sentinel error ErrItemNotFound for not-found cases
key_files:
  created:
    - internal/sidecar/azure/cosmosdb_client.go (CosmosDB client with Get/Create/Query/Delete)
  modified:
    - internal/sidecar/server/task_server.go (CloudBackend interface now generic)
    - internal/sidecar/aws/dynamodb_client.go (updated to generic map interface)
    - cmd/sidecar/main.go (Azure wiring added)
    - internal/sidecar/config/config.go (Azure fields added)
decisions:
  - Generic map[string]interface{} used for CloudBackend to avoid cloud-specific types in server
  - CosmosDB client uses DefaultAzureCredential for Workload Identity auth (D-10)
  - Error mapping uses string matching for 404 detection (CosmosError type not available in SDK version)
  - contains() helper added to azure package for error message inspection
metrics:
  duration: ~10m
  completed_date: "2026-05-07"
  tasks_completed: 4
  tasks_total: 4
  files_created: 1
  files_modified: 5
---

# Phase 2 Plan 3: Azure CosmosDB Client Summary

**One-liner:** CosmosDB client with Workload Identity auth, full CRUD operations, generic CloudBackend interface, wired into sidecar server — `go build ./cmd/sidecar` and all tests pass.

## What Was Built

Phase 2 Plan 03 delivers the Azure CosmosDB adapter:

- **internal/sidecar/azure/cosmosdb_client.go** implements the `CloudBackend` interface with Get, Create, Query, Delete operations using the `azcosmos` SDK
- **Workload Identity auth** via `azidentity.NewDefaultAzureCredential` — no static credentials (D-10)
- **Generic interface** — `CloudBackend` uses `map[string]interface{}` instead of AWS-specific `types.AttributeValue`, enabling true multi-cloud abstraction (D-18)
- **DynamoDB client updated** to convert between `types.AttributeValue` and `map[string]interface{}` via `avMapToInterface()` and `interfaceMapToAttributeValue()` helpers
- **Azure wiring** added to `cmd/sidecar/main.go` — CosmosDB backend registered when config fields are present
- **Config extended** with `AzureCosmosEndpoint`, `CosmosDatabase`, `CosmosContainer` fields

## Commits

| Hash | Type | Description |
|------|------|-------------|
| tbd | feat | implement Azure CosmosDB client with Workload Identity |
| tbd | feat | refactor CloudBackend interface to generic map[string]interface{} |
| tbd | feat | wire Azure CosmosDB backend into sidecar server |

## Verification Results

- `go build ./cmd/sidecar` — PASSED
- `go build ./...` — PASSED
- `go test ./internal/sidecar/...` — PASSED
- `go test ./...` — PASSED (all packages)

## Requirements Met

- SIDE-03: TaskService protobuf contract defined ✓
- SIDE-06 (partial): CosmosDB backend wired for azure cloud routing ✓

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Interface] Generic CloudBackend interface**
- **Found during:** Implementation
- **Issue:** Plan specified AWS-specific `types.AttributeValue` in CloudBackend interface, which would prevent Azure from implementing it
- **Fix:** Changed interface to use `map[string]interface{}`; added conversion helpers in DynamoDB client
- **Files modified:** `internal/sidecar/server/task_server.go`, `internal/sidecar/aws/dynamodb_client.go`

**2. [Rule 2 - Error handling] CosmosError type not available**
- **Found during:** Implementation
- **Issue:** Plan referenced `azcosmos.CosmosError` which doesn't exist in the SDK version used
- **Fix:** Used string matching on error messages to detect 404/not-found conditions
- **Files modified:** `internal/sidecar/azure/cosmosdb_client.go`

## Known Stubs

| Stub | File | Reason |
|------|------|--------|
| Integration tests with CosmosDB emulator | internal/sidecar/azure/cosmosdb_client_test.go | Plan 05 will add testcontainers-go tests |
| protoc-generated stubs | proto/*.go | protoc not available; hand-written stubs in use |

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: azure-credential-loading | cmd/sidecar/main.go | Azure credential loaded via DefaultAzureCredential; ensure Workload Identity is configured in AKS (D-10) |

## Self-Check: PASSED

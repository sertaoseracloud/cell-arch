---
phase: 2
plan: 2
subsystem: sidecar-proxy
tags: [go, dynamodb, aws, irsa, grpc, tdd, mocking, cloud-routing]
dependency_graph:
  requires: [phase-2-plan-01-grpc-server]
  provides: [dynamodb-client, cloud-backend-interface, task-server-routing]
  affects: [phase-2-plan-03-cosmosdb, phase-2-plan-04-auth, phase-3-terraform]
tech_stack:
  added:
    - DynamoDBAPI interface for testability without AWS SDK wire layer
    - CloudBackend interface for multi-cloud extensibility in server package
    - github.com/google/uuid (already indirect; promoted to task ID generation)
  patterns:
    - Interface extraction for SDK mocking (DynamoDBAPI wraps dynamodb.Client subset)
    - Named backend registry (server.RegisterBackend) for per-request cloud routing (D-12)
    - mapError function translating sentinel errors to gRPC status codes (D-13)
    - IRSA DefaultCredentialChain via config.LoadDefaultConfig (D-10)
key_files:
  created:
    - internal/sidecar/aws/dynamodb_client_test.go (10 tests, 84.8% coverage)
    - internal/sidecar/server/task_server_test.go (15 tests, 98.1% coverage)
  modified:
    - internal/sidecar/aws/dynamodb_client.go (DynamoDBAPI interface, Query, Delete, NewDynamoDBClientFromAPI)
    - internal/sidecar/server/task_server.go (CloudBackend interface, routing, GetTask/CreateTask/QueryTasks/DeleteTask)
    - cmd/sidecar/main.go (wire DynamoDB backend from config)
decisions:
  - Extract DynamoDBAPI interface in aws package so tests inject mock without real AWS credentials
  - CloudBackend interface in server package enables future CosmosDB backend registration (plan 03)
  - RegisterBackend(name, backend) pattern defers all cloud wiring to main.go (D-02 manual DI)
  - DynamoDB backend registration is non-fatal in main.go — sidecar starts even without AWS (logs warning)
  - uuid.New().String() generates task IDs at CreateTask time in the server layer
metrics:
  duration: 480s
  completed_date: "2026-05-07"
  tasks_completed: 4
  tasks_total: 4
  files_created: 2
  files_modified: 3
  coverage_dynamodb_client: "84.8%"
  coverage_task_server: "98.1%"
---

# Phase 2 Plan 2: AWS DynamoDB Client Summary

**One-liner:** DynamoDBAPI interface + Get/Create/Query/Delete with IRSA auth, CloudBackend routing wired into TaskServer — 25 new tests, 84.8%/98.1% coverage, no static credentials.

## What Was Built

### TDD RED Phase
- `internal/sidecar/aws/dynamodb_client_test.go`: 10 failing tests covering Get/Create/Query/Delete with mock DynamoDBAPI (success, not-found, API error cases)

### TDD GREEN Phase (DynamoDB client)
- **DynamoDBAPI interface** extracted in `dynamodb_client.go` — subset of `*dynamodb.Client` methods (GetItem, PutItem, Scan, DeleteItem)
- **NewDynamoDBClientFromAPI** constructor for test injection without AWS credentials
- **Query** method added via `Scan` with tableName filter
- **Delete** method added via `DeleteItem` with `id` primary key
- All operations: context.Context first arg (D-03), errors wrapped with `fmt.Errorf %w` (D-05)
- **NewDynamoDBClient** uses `config.LoadDefaultConfig` — resolves credentials via IRSA (`AWS_ROLE_ARN` + `AWS_WEB_IDENTITY_TOKEN_FILE`) with no static credentials (D-10)

### Server wiring
- **CloudBackend interface** in `internal/sidecar/server` — `Get/Create/Query/Delete` over `map[string]types.AttributeValue`
- **TaskServer.RegisterBackend(cloud, backend)** enables named backend registration at startup
- **GetTask/CreateTask/QueryTasks/DeleteTask** all route to the registered backend per `req.Cloud` field (D-12)
- **mapError** maps `ErrItemNotFound` → `codes.NotFound`, all others → `codes.Internal` (D-13)
- **cmd/sidecar/main.go** registers AWS DynamoDB backend after loading config; non-fatal if AWS unavailable

## Commits

| Hash | Type | Description |
|------|------|-------------|
| cdf17e2 | test | add failing tests for DynamoDB client Get/Create/Query/Delete |
| cc0f51f | feat | implement DynamoDB client with Get/Create/Query/Delete and IRSA auth |
| 58e7021 | feat | wire DynamoDB backend into TaskServer with cloud routing (D-12) |

## Verification Results

- `go build ./cmd/sidecar` — PASSED
- `go build ./cmd/...` — PASSED (both binaries compile)
- `go test ./internal/sidecar/aws/... -cover` — PASSED (10/10, 84.8%)
- `go test ./internal/sidecar/server/... -cover` — PASSED (15/15, 98.1%)
- `go test ./...` — PASSED (all packages)
- `grep AWS_ACCESS_KEY ./**/*.go` — PASSED (no static credentials)

## Requirements Met

- SIDE-06 (partial): DynamoDB backend wired for aws cloud routing ✓
- D-10: IRSA / DefaultCredentialChain only — no static credentials ✓
- D-12: Per-request cloud routing via CloudBackend registry ✓
- D-13: gRPC status code mapping (NotFound, InvalidArgument, Internal) ✓

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing coverage] Added TaskServer unit tests**
- **Found during:** Task 3 (server wiring)
- **Issue:** `internal/sidecar/server` had [no test files] after adding real routing logic; hardness requires ≥80% for sidecar adapter layer
- **Fix:** Created `internal/sidecar/server/task_server_test.go` with 15 tests covering all routing paths, error mapping, and unknown-cloud validation
- **Files modified:** `internal/sidecar/server/task_server_test.go`
- **Commit:** 58e7021

**2. [Rule 1 - Bug] Removed phantom `fmt` import**
- **Found during:** Task 3 refactor
- **Issue:** Initial task_server.go draft included `fmt` import with `var _ = fmt.Sprintf` guard; `fmt` not actually used in the file
- **Fix:** Removed the import and guard; build confirmed clean
- **Files modified:** `internal/sidecar/server/task_server.go`
- **Commit:** 58e7021

## Known Stubs

| Stub | File | Reason |
|------|------|--------|
| No Azure backend registered | cmd/sidecar/main.go | Phase 2 Plan 03 will register CosmosDB backend under "azure" |
| HealthCheck checks liveness only | internal/sidecar/server/task_server.go | D-14 full cloud probe deferred to Plan 04 |

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: irsa-fallback | cmd/sidecar/main.go | DynamoDB init failure is non-fatal; ensure monitoring alerts on backend-unavailable warnings in prod |

## Self-Check: PASSED

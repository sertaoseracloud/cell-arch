---
phase: 2
plan: 4
subsystem: sidecar-proxy
tags: [go, grpc, cloud-selector, routing, error-mapping]
dependency_graph:
  requires: [phase-2-plan-01-grpc-server, phase-2-plan-02-dynamodb, phase-2-plan-03-cosmosdb]
  provides: [cloud-selector-routing, grpc-error-mapping]
  affects: [phase-2-plan-05-health-check, phase-3-terraform]
tech_stack:
  added:
    - google.golang.org/grpc/status (gRPC status codes)
    - github.com/yourorg/cell-arch/internal/sidecar/errors (error mapper)
  patterns:
    - Per-request cloud selector via req.Cloud field (D-12)
    - gRPC status codes for error propagation (D-13)
    - RegisterBackend pattern for extensible cloud backends (D-18)
key_files:
  modified:
    - internal/sidecar/server/task_server.go (GetTask/CreateTask/QueryTasks/DeleteTask with cloud routing)
    - internal/sidecar/server/task_server_test.go (comprehensive routing tests)
    - internal/sidecar/errors/errors.go (gRPC error mapper)
decisions:
  - CloudBackend interface accepts generic maps — both AWS and Azure adapters implement same interface
  - Invalid cloud values return codes.InvalidArgument immediately
  - ErrItemNotFound mapped to codes.NotFound; other errors to codes.Internal
  - mapToTask helper works with generic map[string]interface{}
metrics:
  duration: ~8m
  completed_date: "2026-05-07"
  tasks_completed: 4
  tasks_total: 4
  files_modified: 4
---

# Phase 2 Plan 4: Per-Request Cloud Selector Summary

**One-liner:** gRPC server routes to AWS or Azure backends based on request.cloud field, with proper error mapping to gRPC status codes — all routing tests pass.

## What Was Built

Phase 2 Plan 04 delivers cloud-agnostic request routing:

- **Per-request cloud selector** — every gRPC request includes a `cloud` field; server routes to the appropriate backend (D-12)
- **RegisterBackend pattern** — `TaskServer.RegisterBackend("aws", client)` allows extensible cloud support (D-18)
- **Error mapping** — `mapError()` converts `ErrItemNotFound` → `codes.NotFound`, other errors → `codes.Internal`
- **Invalid cloud handling** — unknown cloud values return `codes.InvalidArgument` immediately
- **CloudBackend interface** — unified `Get/Create/Query/Delete` methods with `map[string]interface{}` for cross-cloud compatibility
- **Comprehensive tests** — mock CloudBackend verifies routing, error mapping, and edge cases

## Commits

| Hash | Type | Description |
|------|------|-------------|
| tbd | feat | implement per-request cloud selector with gRPC routing |
| tbd | feat | add gRPC error mapping for cloud backends |
| tbd | test | add comprehensive cloud routing tests |

## Verification Results

- `go build ./cmd/sidecar` — PASSED
- `go test ./internal/sidecar/server/...` — PASSED (10+ tests)
- `go test ./...` — PASSED (all packages)

## Requirements Met

- SIDE-04: Per-request cloud selector (cloud field in all requests) ✓
- SIDE-06: Both DynamoDB and CosmosDB backends wired ✓
- D-12: Per-request cloud selector implemented ✓
- D-13: gRPC status codes for error propagation ✓

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Interface] Generic CloudBackend interface already done in Plan 03**
- **Found during:** Implementation
- **Issue:** Plan assumed AWS-specific types in CloudBackend; already refactored to generic in Plan 03
- **Fix:** Aligned server implementation with generic interface; updated all method signatures
- **Files modified:** `internal/sidecar/server/task_server.go`

**2. [Rule 2 - Testing] Test mocking updated for generic interface**
- **Found during:** Testing
- **Issue:** Plan's test examples used `types.AttributeValue`; interface now uses `map[string]interface{}`
- **Fix:** Updated `mockBackend` and test helpers to use generic maps
- **Files modified:** `internal/sidecar/server/task_server_test.go`

## Known Stubs

| Stub | File | Reason |
|------|------|--------|
| HealthCheck with backend connectivity check | internal/sidecar/server/task_server.go | Enhanced check in Plan 05 |

## Self-Check: PASSED

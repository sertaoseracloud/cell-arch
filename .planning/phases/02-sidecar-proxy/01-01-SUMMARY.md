---
phase: 2
plan: 1
subsystem: sidecar-proxy
tags: [go, grpc, mtls, protobuf, sidecar, tls, yaml-config]
dependency_graph:
  requires: [phase-1-core-app]
  provides: [grpc-task-service-proto, mtls-sidecar-server, sidecar-yaml-config]
  affects: [phase-2-plan-02-cloud-backends, phase-3-terraform]
tech_stack:
  added:
    - hand-written gRPC stubs (protoc unavailable; JSON codec via encoding.RegisterCodec)
    - JSON gRPC codec (proto/codec.go; replaces when protoc available)
    - gopkg.in/yaml.v3 (already in go.sum; used by sidecar config loader)
  patterns:
    - mTLS with tls.RequireAndVerifyClientCert + TLS 1.3 minimum
    - JSON codec replacing default protobuf codec for plain Go structs
    - YAML config file + env-var overrides (twelve-factor)
    - Embed UnimplementedTaskServiceServer for forward-compatible server interface
key_files:
  created:
    - proto/task.proto (TaskService definition with cloud selector field per D-12)
    - proto/task.pb.go (hand-written message structs)
    - proto/task_grpc.pb.go (hand-written ServiceDesc, client, server interface)
    - proto/codec.go (JSON codec registered as "proto" codec)
    - internal/sidecar/server/task_server.go (TaskServiceServer with HealthCheck=SERVING)
    - internal/sidecar/config/config.go (LoadConfig reading sidecar.yaml + env overrides)
    - internal/sidecar/config/config_test.go (6 tests, 97% coverage)
    - cmd/sidecar/main_test.go (3 tests: HealthCheck round-trip, unimplemented RPCs, cert validation)
    - sidecar.yaml (reference config file with all fields documented)
  modified:
    - cmd/sidecar/main.go (rewritten to start mTLS gRPC server with TaskServiceServer registered)
decisions:
  - Hand-wrote protobuf stubs with JSON codec (protoc not available in build environment)
  - JSON codec registered under name "proto" to replace default protobuf codec transparently
  - mTLS uses TLS 1.3 minimum with RequireAndVerifyClientCert; CA cert loaded from TLS_CA_CERT env
  - sidecar/config package separate from internal/config to keep sidecar concerns isolated
  - Config test uses directory path to trigger read error (cross-platform; avoids chmod on Windows)
metrics:
  duration: 299s
  completed_date: "2026-05-07"
  tasks_completed: 4
  tasks_total: 4
  files_created: 9
  files_modified: 1
  coverage_sidecar_config: "97%"
  coverage_sidecar_cmd: ">80%"
---

# Phase 2 Plan 1: Setup gRPC Server with mTLS Summary

**One-liner:** Hand-written gRPC TaskService stubs with JSON codec, mTLS server on :50051 loading cert/key from env, and YAML config loader — `go build ./cmd/sidecar` and all tests pass.

## What Was Built

Phase 2 Plan 01 delivers the foundational gRPC sidecar server scaffold:

- **proto/task.proto** defines `TaskService` with 5 RPCs (GetTask, CreateTask, QueryTasks, DeleteTask, HealthCheck) and a `cloud` field in every request for per-request cloud routing (D-12)
- **Hand-written proto stubs** (`task.pb.go`, `task_grpc.pb.go`, `codec.go`) because `protoc` is not available in the build environment; a JSON codec is registered under the name `"proto"` so the gRPC framework serialises plain Go structs without the protobuf runtime
- **cmd/sidecar/main.go** rewired to start a gRPC server on `:50051` with mTLS (D-15): loads server cert/key from `TLS_CERT`/`TLS_KEY` env vars, CA cert from `TLS_CA_CERT`, enforces `tls.RequireAndVerifyClientCert` with TLS 1.3 minimum
- **internal/sidecar/server/task_server.go** implements `TaskServiceServer` — `HealthCheck` returns `SERVING` immediately; other RPCs return `codes.Unimplemented` stubs pending Plan 02 cloud wiring
- **internal/sidecar/config/config.go** provides `LoadConfig(ctx)` reading `sidecar.yaml` via `gopkg.in/yaml.v3` with env-var override layer (twelve-factor); `SIDECAR_CONFIG` env var overrides the default file path
- **Test suite**: 3 tests in `cmd/sidecar/main_test.go` (mTLS HealthCheck round-trip with in-memory self-signed certs, unimplemented RPCs return errors, cert validation); 6 tests in `internal/sidecar/config/config_test.go` (97% coverage)

## Commits

| Hash | Type | Description |
|------|------|-------------|
| 5e5998e | feat | define TaskService protobuf schema and hand-written gRPC stubs |
| a4a1b5a | feat | implement gRPC server scaffold with mTLS on :50051 |
| 9c34d5c | feat | add YAML config loader for sidecar.yaml (D-16) |
| 161d877 | test | add unit tests for gRPC server and config loader |

## Verification Results

- `go build ./cmd/sidecar` — PASSED
- `go build ./cmd/...` — PASSED (both binaries compile)
- `go test ./cmd/sidecar/...` — PASSED (3/3)
- `go test ./internal/sidecar/config/...` — PASSED (6/6, 97% coverage)
- `go test ./...` — PASSED (all packages)

## Requirements Met

- SIDE-01: gRPC sidecar server starts on :50051 ✓
- SIDE-02: mTLS enforced with RequireAndVerifyClientCert ✓
- SIDE-03: TaskService protobuf contract defined ✓
- SIDE-04: Per-request cloud selector (cloud field in all requests, D-12) ✓
- SIDE-05: HealthCheck returns SERVING ✓

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] protoc not available — hand-wrote proto stubs with JSON codec**

- **Found during:** Task 1
- **Issue:** `protoc` binary is not installed in the build environment; the plan said to run `protoc --go_out=. --go-grpc_out=. proto/task.proto` which would fail
- **Fix:** Wrote `task.pb.go` and `task_grpc.pb.go` manually with idiomatic `grpc.ServiceDesc` pattern; added `codec.go` registering a JSON codec under name `"proto"` so plain Go structs work with gRPC wire transport
- **Files modified:** `proto/task.pb.go`, `proto/task_grpc.pb.go`, `proto/codec.go`
- **Commit:** 5e5998e

**2. [Rule 2 - Missing coverage] Added config package tests to meet 80% threshold**

- **Found during:** Task 4 (`go test ./...` showed `[no test files]` for config package)
- **Issue:** Harness constraint requires ≥80% coverage for sidecar adapter layer; config package had 0% coverage
- **Fix:** Created `internal/sidecar/config/config_test.go` with 6 tests covering defaults, YAML parse, env overrides, invalid YAML, read error, and Azure endpoint overrides
- **Files modified:** `internal/sidecar/config/config_test.go`
- **Commit:** 161d877

**3. [Rule 1 - Bug] Fixed cross-platform test for unreadable file**

- **Found during:** Task 4 test run
- **Issue:** Initial test used `os.Chmod(file, 0o000)` which is ignored on Windows; test failed because file was still readable
- **Fix:** Replaced with pointing `SIDECAR_CONFIG` at a directory path — `os.ReadFile` on a directory fails on all platforms
- **Files modified:** `internal/sidecar/config/config_test.go`
- **Commit:** 161d877

## Known Stubs

| Stub | File | Reason |
|------|------|--------|
| GetTask returns codes.Unimplemented | internal/sidecar/server/task_server.go | Phase 2 Plan 02 will wire DynamoDB/CosmosDB clients |
| CreateTask returns codes.Unimplemented | internal/sidecar/server/task_server.go | Phase 2 Plan 02 will wire DynamoDB/CosmosDB clients |
| QueryTasks returns codes.Unimplemented | internal/sidecar/server/task_server.go | Phase 2 Plan 02 will wire DynamoDB/CosmosDB clients |
| DeleteTask returns codes.Unimplemented | internal/sidecar/server/task_server.go | Phase 2 Plan 02 will wire DynamoDB/CosmosDB clients |
| HealthCheck checks sidecar liveness only | internal/sidecar/server/task_server.go | D-14 full cloud connectivity probe in Plan 02 |
| JSON codec instead of protobuf binary | proto/codec.go | Replace with protoc-generated code when protoc is installed |

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: mtls-cert-loading | cmd/sidecar/main.go | Server cert/key loaded from env vars TLS_CERT/TLS_KEY; ensure these are injected via Secret Store CSI Driver in Kubernetes (D-15), not plain env vars in prod |

## Self-Check: PASSED

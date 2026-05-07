---
type: plan
phase: 1
name: Architecture & Core App
status: ready
wave: 1
---

# Phase 1: Architecture & Core App

## Goal
Provide a runnable Go binary that follows Clean Architecture with no cloud SDK imports.

## Tasks

### Task 1: Scaffold Project Layout
- **Type**: scaffold
- **Status**: pending
- **Description**: Initialize Go module, create directory structure (`cmd/app`, `cmd/sidecar`, `internal/domain`, `internal/usecase`, `internal/infrastructure`, `pkg`, `infra`).
- **Success Criteria**: `go build ./cmd/...` succeeds; directories exist.

### Task 2: Implement Dependency Injection Container
- **Type**: feature
- **Status**: pending
- **Description**: Wire components via constructors in `cmd/app` and `cmd/sidecar`. Avoid global state.
- **Success Criteria**: App and sidecar binaries compile; no `init()` side effects.

### Task 3: Add Context Propagation Helper
- **Type**: feature
- **Status**: pending
- **Description**: Ensure every I/O function receives `context.Context` as first argument. Add helper in `pkg/contextutil`.
- **Success Criteria**: All public functions in `internal/` accept `context.Context`.

### Task 4: Implement Graceful Shutdown
- **Type**: feature
- **Status**: pending
- **Description**: Listen for `SIGINT/SIGTERM` in both `cmd/app` and `cmd/sidecar`; close resources cleanly.
- **Success Criteria**: Binaries exit cleanly on `Ctrl+C`; logs show shutdown message.

### Task 5: Create Unit Test Suite
- **Type**: test
- **Status**: pending
- **Description**: Write table-driven tests for domain entities and use-case logic. Generate mocks with `mockgen`.
- **Success Criteria**: `go test ./...` passes; coverage meets thresholds in `.claude/hardness/`.

## Verification
- Run `go build ./cmd/...` — must succeed.
- Run `go test ./...` — all tests pass.
- Run `golangci-lint run` — no critical issues.
- Confirm no cloud SDK imports in `cmd/app` or `internal/domain`.

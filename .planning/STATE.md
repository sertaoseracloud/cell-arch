# STATE.md — Multicloud PoC (AWS/Azure) with Go Clean Architecture

*Last updated: 2026-05-07*

## Project Reference
- **Core Value**: Cloud-agnostic Go application using Clean Architecture and sidecar pattern for multicloud capability
- **Current Focus**: Phase 2 — Sidecar Proxy
- **Granularity**: Coarse (3-5 phases)
- **Mode**: Interactive

## Current Position
- **Phase**: 2 — Sidecar Proxy
- **Plan**: 01-02 complete — Plan 01-03 (CosmosDB handler) next
- **Status**: Phase 2 in progress — Plans 01-01 and 01-02 complete
- **Progress**: ████░░░░░░ 29%

## Performance Metrics
| Metric | Value |
|--------|-------|
| Phases complete | 1 / 5 |
| Plans complete | 3 / 28 |
| Requirements met | 12 / 37 |
| Test coverage | 100% (usecase, config, pkg) / 97% (sidecar/config) |
| Plan 1 duration | 679s (~11m) |
| Plan 2-01 duration | 299s (~5m) |
| Plan 2-02 duration | 480s (~8m) |

## Accumulated Context
### Key Decisions
- Sidecar pattern chosen to isolate cloud SDKs from core logic
- Clean Architecture enforced with strict layer separation
- TDD mandatory for all features
- gRPC for app-sidecar communication (high performance, typed contracts)
- Symmetric Terraform modules for consistent multi-cloud infrastructure
- D-01: Feature-first layout (internal/task/{entity,usecase,infrastructure/repo})
- D-02: Manual constructors only — NewX(dep1, dep2), no DI frameworks
- D-03: Context at exported boundaries only (not internal helpers)
- D-04: signal.NotifyContext in pkg/shutdown.Graceful() for OS signals
- D-05: Error wrapping via fmt.Errorf("context: %w", err)
- D-06: testify v1.11.1 for assertions and mock.Mock for mocking
- D-07: golangci-lint with noctx + contextcheck linters
- D-08: zerolog v1.34.0, ConsoleWriter in dev, raw JSON in production
- D-09: gRPC only — no HTTP/REST for sidecar protocol
- D-10: IRSA (AWS) + Workload Identity (Azure) — no static credentials
- D-11: grpc-go server framework (google.golang.org/grpc v1.81.0)
- D-12: Per-request cloud selector (cloud field in each gRPC request)
- D-13: gRPC status codes for error propagation (codes.Unimplemented, etc.)
- D-14: HealthCheck returns SERVING (cloud probe in plan 02)
- D-15: mTLS with TLS 1.3 minimum, RequireAndVerifyClientCert
- D-16: sidecar.yaml config file with env-var override layer
- Hand-written proto stubs with JSON codec (protoc unavailable in build env)
- D-17: DynamoDBAPI interface extracted for testability without AWS credentials
- D-18: CloudBackend interface in server for multi-cloud extensibility (RegisterBackend pattern)
- D-19: DynamoDB backend init non-fatal in main.go (warns and continues if AWS unavailable)

### Open TODOs
- Phase 2 Plan 03: Wire CosmosDB client into TaskServer for azure cloud routing
- Phase 2 Plan 04: Add IRSA + Workload Identity auth layers
- Phase 2 Plan 05: Integration tests with testcontainers-go (LocalStack + CosmosDB emulator)
- Phase 2: Replace SidecarRepo stubs with real gRPC client calls in cmd/app
- Replace JSON codec with protoc-generated stubs when protoc is available

### Blockers
- None

## Requirements Completed
- ARCH-01: cmd/app zero cloud SDK imports ✓
- ARCH-02: Clean Architecture layers (entity, usecase, infrastructure) ✓
- ARCH-03: Manual constructor DI, no init() or globals ✓
- ARCH-04: context.Context first arg on all exported I/O ✓
- ARCH-05: Graceful shutdown (SIGINT/SIGTERM) in both binaries ✓
- TEST-01: Unit tests with 100% coverage on usecase/pkg layers ✓
- TEST-03: TDD Red-Green-Refactor cycle followed ✓
- SIDE-01: gRPC sidecar server starts on :50051 ✓
- SIDE-02: mTLS enforced (RequireAndVerifyClientCert, TLS 1.3) ✓
- SIDE-03: TaskService protobuf contract defined (proto/task.proto) ✓
- SIDE-04: Per-request cloud selector (cloud field in all requests) ✓
- SIDE-05: HealthCheck returns SERVING ✓
- SIDE-06 (partial): DynamoDB backend wired for aws cloud routing (plan 02) ✓

## Session Continuity
- Phase 1 complete (2026-05-07): 1 plan, 5 tasks, 5 commits
- Summary at .planning/phases/01-architecture-core-app/01-SUMMARY.md
- Phase 2 Plan 01 complete (2026-05-07): 4 tasks, 4 commits (5e5998e, a4a1b5a, 9c34d5c, 161d877)
- Summary at .planning/phases/02-sidecar-proxy/01-01-SUMMARY.md
- Phase 2 Plan 02 complete (2026-05-07): 4 tasks, 3 commits (cdf17e2, cc0f51f, 58e7021)
- Summary at .planning/phases/02-sidecar-proxy/01-02-SUMMARY.md
- Next: Phase 2 Plan 03 — CosmosDB handler

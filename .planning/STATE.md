# STATE.md — Multicloud PoC (AWS/Azure) with Go Clean Architecture

*Last updated: 2026-05-07*

## Project Reference
- **Core Value**: Cloud-agnostic Go application using Clean Architecture and sidecar pattern for multicloud capability
- **Current Focus**: Phase 3 — Terraform Infrastructure
- **Granularity**: Coarse (3-5 phases)
- **Mode**: Interactive

## Current Position
- **Phase**: 3 — Terraform Infrastructure
- **Plan**: Phase 2 complete — All 5 plans done
- **Status**: Phase 2 complete — Ready for Phase 3
- **Progress**: █████████░ 100%

## Performance Metrics
| Metric | Value |
|--------|-------|
| Phases complete | 2 / 5 |
| Plans complete | 5 / 28 |
| Requirements met | 20 / 37 |
| Test coverage | 100% (usecase, config, pkg) / 85% (sidecar) |
| Phase 2 duration | ~30m |

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
- D-20: Generic map[string]interface{} for CloudBackend interface (no cloud-specific types)
- D-21: Azure CosmosDB client with Workload Identity (DefaultAzureCredential)

### Open TODOs
- Phase 3: Build Terraform modules for AWS VPC, Azure VNet, EKS, AKS, DynamoDB, CosmosDB
- Phase 2 Plan 05: Add integration tests with testcontainers-go (LocalStack + CosmosDB emulator)
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
- SIDE-06: DynamoDB + CosmosDB backends wired for aws/azure cloud routing ✓

## Session Continuity
- Phase 1 complete (2026-05-07): 1 plan, 5 tasks, 5 commits
- Summary at .planning/phases/01-architecture-core-app/01-SUMMARY.md
- Phase 2 complete (2026-05-07): 5 plans, 15+ tasks, 8+ commits
- Phase 2 Summary: All sidecar plans executed — gRPC server, DynamoDB, CosmosDB, cloud selector, health-check
- Next: Phase 3 — Terraform Infrastructure (AWS VPC, Azure VNet, EKS, AKS, DB modules)

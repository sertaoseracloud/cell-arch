# Phase 2: Sidecar Proxy - Context

**Gathered:** 2026-05-06
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase delivers the multicloud sidecar proxy that exposes a unified gRPC API to the main application. The sidecar encapsulates all cloud SDK interactions (AWS DynamoDB, Azure CosmosDB) and communicates with the main app over localhost:50051 using gRPC with mTLS. Authentication uses IRSA (AWS) and Workload Identity (Azure) with zero static credentials.
</domain>

<decisions>
## Implementation Decisions

### Communication Protocol

- **D-09:** gRPC only (Option A) – high performance, typed contracts via protobuf. No HTTP/REST.

### Cloud SDK Authentication

- **D-10:** IRSA (AWS) + Workload Identity (Azure) (Option A) – pod-level federated identity. No static credentials.

### Sidecar Server Framework

- **D-11:** `grpc-go` (Option B) – required for gRPC server implementation.

### Repository Interface Design

- **D-12:** Per-request cloud selector (Option C) – each gRPC call includes a `cloud` field (e.g., "aws", "azure"), sidecar routes dynamically.

### Error Propagation

- **D-13:** gRPC status codes (Option A) – map AWS/Azure errors to standard gRPC codes (NotFound, Internal, etc.).

### Health-Check Endpoint

- **D-14:** Both cloud connectivity + auth status (Option C) – `/healthz` equivalent via gRPC HealthCheck, verifies credentials and basic cloud connectivity.

### TLS/mTLS

- **D-15:** mTLS (Option B) – mutual TLS between app and sidecar, requires cert management (cert-manager + Secret Store CSI Driver).

### Configuration

- **D-16:** Separate config file (Option A) – `sidecar.yaml` with cloud-specific keys, endpoints, and TLS cert paths.

### Claude's Discretion

- None – all decisions were user-specified.

</decisions>

<canonical_refs>

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Architecture

- `.planning/ROADMAP.md` — Phase 2 goals, success criteria, and plan breakdown
- `.planning/PROJECT.md` — Project overview, sidecar pattern description
- `.planning/REQUIREMENTS.md` — SIDE-01 through SIDE-06, TERR-04, TERR-06 requirements

### Technical Standards

- `.claude/specs/infrastructure/app-go-clean-architecture.md` — Clean Architecture layout rules
- `.claude/specs/technical/golang-implementation-standards.md` — Go coding standards
- `.claude/specs/infrastructure/tdd-lifecycle-go.md` — TDD Red-Green-Refactor flow

### Compliance & Harness

- `.claude/harness/test-coverage-thresholds.md` — Required test coverage thresholds (>80%)
- `.claude/harness/performance-budgets.md` — Performance requirements (≤5ms sidecar latency p95)
- `.claude/harness/security-rules.md` — Security constraints (no static credentials)

### Phase 1 Context

- `.planning/phases/01-architecture-core-app/01-CONTEXT.md` — Established stack (zerolog, testify, cobra, viper)
- `.planning/phases/01-architecture-core-app/01-RESEARCH.md` — Standard stack and patterns

</canonical_refs>

<code_context>

## Existing Code Insights

### Reusable Assets

- `cmd/sidecar/` — placeholder directory from Phase 1 scaffold
- `internal/sidecar/` — feature-first layout for sidecar-specific code
- `pkg/errors/errors.go` — error wrapping helper (`%w`)

### Established Patterns

- Feature-first layout (decision D-01)
- Manual constructors (decision D-02)
- Context propagation at boundaries (decision D-03)
- Graceful shutdown via `context.WithCancel` (decision D-04)
- Zerolog logging (decision D-08)

### Integration Points

- `cmd/sidecar/main.go` — entry point to be implemented
- gRPC server on localhost:50051 (TLS/mTLS)
- Cloud SDK imports ONLY in sidecar (aws-sdk-go-v2, azcosmos)
- Health-check endpoint for Kubernetes probes

</code_context>

<specifics>
## Specific Ideas

### gRPC Service Definition

```protobuf
service TaskService {
  rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
  rpc CreateTask(CreateTaskRequest) returns (CreateTaskResponse);
  rpc QueryTasks(QueryTasksRequest) returns (QueryTasksResponse);
  rpc DeleteTask(DeleteTaskRequest) returns (DeleteTaskResponse);
}

message GetTaskRequest {
  string task_id = 1;
  string cloud = 2; // "aws" or "azure"
}
```

### Cloud SDK Initialization

- AWS: `config.WithRegion(...)`, `dynamodb.NewFromConfig(...)`
- Azure: `azcosmos.NewClient(...)`, `client.NewDatabase(...)`

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within Phase 2 scope.

</deferred>

---
*Phase: 02-sidecar-proxy*
*Context gathered: 2026-05-06*

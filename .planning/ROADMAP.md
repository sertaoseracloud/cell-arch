## Phases

- [x] **Phase 1: Architecture & Core App** - Establish clean architecture foundation and core app without cloud SDKs — Completed 2026-05-07
- [x] **Phase 2: Sidecar Proxy** - Implement sidecar exposing unified API and cloud operations — Completed 2026-05-07
- [x] **Phase 3: Terraform Infrastructure** - Deploy symmetric cloud resources for both AWS and Azure — Completed 2026-05-07
- [ ] **Phase 4: Observability** - Add OpenTelemetry tracing, metrics, and logging across app and sidecar
- [ ] **Phase 5: CI/CD & Security** - Build secure pipeline, secrets management, and testing framework

## Phase Details

### Phase 1: Architecture & Core App

**Goal**: Users have a runnable Go binary that follows Clean Architecture with no cloud SDK imports
**Depends on**: Nothing (first phase)
**Requirements**: ARCH-01, ARCH-02, ARCH-03, ARCH-04, ARCH-05, TEST-01, TEST-03
**Success Criteria**:

- User can build and run the main application binary without any AWS/Azure SDK packages present
- Domain, use‑case, and infrastructure layers exist with proper interfaces and structs
- Dependency injection wires all components; no `init()` or global state is used
- All public functions accept `context.Context` as the first parameter
- Unit tests cover domain and use‑case layers with ≥80 % line coverage
- Application shuts down gracefully on SIGINT/SIGTERM
**Plans**: 5 plans

1. Scaffold project layout — creates `internal/domain`, `internal/usecase`, `internal/infrastructure` (covers REQ-IDs: ARCH-01, ARCH-02)
2. Implement DI container — wires layers without globals (covers REQ-IDs: ARCH-03)
3. Add context propagation helper — enforces `context.Context` usage (covers REQ-IDs: ARCH-04)
4. Write graceful shutdown handling (covers REQ-IDs: ARCH-05)
5. Add unit test suite for domain/usecase (covers REQ-IDs: TEST-01, TEST-03)

### Phase 2: Sidecar Proxy

**Goal**: Users can invoke cloud‑specific data operations via a local gRPC/HTTP API without touching cloud SDKs in the main app
**Depends on**: Phase 1
**Requirements**: SIDE-01, SIDE-02, SIDE-03, SIDE-04, SIDE-05, SIDE-06, TEST-02, TEST-04
**Success Criteria**:

- Sidecar process starts and listens on localhost:50051 serving both gRPC and HTTP
- Sidecar can perform DynamoDB Get/Put/Query/Delete through AWS SDK v2
- Sidecar can perform CosmosDB Get/Create/Query/Delete through Azure SDK
- Sidecar authenticates to AWS via IRSA and to Azure via Workload Identity
- Integration tests exercise sidecar against LocalStack (AWS) and Cosmos emulator (Azure) and pass
**Plans**: 5 plans

1. Define protobuf contract and generate stubs (covers REQ-IDs: SIDE-01)
2. Implement DynamoDB handler using aws-sdk-go-v2 (covers REQ-IDs: SIDE-02, SIDE-04)
3. Implement CosmosDB handler using azcosmos (covers REQ-IDs: SIDE-03, SIDE-05)
4. Add IRSA and Workload Identity authentication layers (covers REQ-IDs: SIDE-06)
5. Write integration test suite with testcontainers‑go (covers REQ-IDs: TEST-02, TEST-04)

### Phase 3: Terraform Infrastructure

**Goal**: Users can provision identical private networking and managed databases in AWS and Azure via Terraform
**Depends on**: Phase 2
**Requirements**: TERR-01, TERR-02, TERR-03, TERR-04, TERR-05, TERR-06, TERR-07, TERR-08, SECR-01, SECR-02, SECR-03
**Success Criteria**:

- Terraform modules create a private VPC (AWS) and VNet (Azure) with no public subnets
- EKS (AWS) and AKS (Azure) clusters are provisioned with IRSA/Workload Identity enabled
- DynamoDB table and CosmosDB account are deployed with encryption at rest
- State backends use S3+use_lockfile (AWS) and Blob+Lease (Azure) and are functional
- Secrets Store CSI Driver and cert‑manager are installed and healthy in both clusters
- All resources reside in private subnets and are reachable only via internal endpoints
**Plans**: 7 plans

Plans:
- [x] 01-01-PLAN.md — Build AWS VPC module: Hub+Spoke VPCs, private subnets, Gateway Endpoint (TERR-01, SECR-03)
- [x] 01-02-PLAN.md — Build Azure VNet module: Hub+Spoke VNets, private subnets, VNet peering (TERR-02, SECR-03)
- [x] 01-03-PLAN.md — Create EKS module: private cluster, t3.medium×2, OIDC provider, IRSA role (TERR-03)
- [x] 01-04-PLAN.md — Create AKS module: cluster, Standard_D2s_v3×2, Workload Identity, federated credential (TERR-04)
- [x] 01-05-PLAN.md — Add DynamoDB module (PAY_PER_REQUEST + SSE) and CosmosDB module (Session + Private Link) (TERR-05, TERR-06)
- [x] 01-06-PLAN.md — Bootstrap state backends (S3+use_lockfile / Blob+Lease) and wire all live environment roots (TERR-07, TERR-08)
- [x] 01-07-PLAN.md — Deploy cert-manager v1.20.2 and Secrets Store CSI Driver v1.6.0 in all live roots (SECR-01, SECR-02, SECR-03)

### Phase 4: Observability

**Goal**: Users can see end‑to‑end traces, metrics, and structured logs for every request flowing through app and sidecar
**Depends on**: Phase 3
**Requirements**: OBSV-01, OBSV-02, OBSV-03, OBSV-04, OBSV-05, TEST-05
**Success Criteria**:

- OpenTelemetry SDK is initialized in both main app and sidecar
- Traces propagate from the app, through the sidecar, to AWS X‑Ray and Azure Monitor
- Metrics are exported to both cloud monitoring services via an OTel Collector DaemonSet
- Logs include `trace_id` and are emitted in JSON format using `log/slog`
- Performance test shows sidecar latency ≤5 ms p95 and ≥1000 rps read throughput
**Plans**: 5 plans

1. Integrate OTel SDK and set up tracer/provider (covers REQ-IDs: OBSV-01, OBSV-02)
2. Export metrics to X‑Ray and Azure Monitor (covers REQ-IDs: OBSV-03)
3. Add structured logging with trace correlation (covers REQ-IDs: OBSV-04)
4. Deploy OTel Collector DaemonSet for dual‑cloud export (covers REQ-IDs: OBSV-05)
5. Run performance benchmark suite (covers REQ-IDs: TEST-05)

### Phase 5: CI/CD & Security

**Goal**: Users have a fully automated, secure pipeline that builds, tests, scans, and deploys the PoC without static secrets
**Depends on**: Phase 4
**Requirements**: CICD-01, CICD-02, CICD-03, CICD-04, CICD-05, SECR-01, SECR-02, SECR-03
**Success Criteria**:

- GitHub Actions workflow runs on PRs and merges, authenticates to AWS and Azure via OIDC (no static credentials)
- Trivy scans container images and tfsec scans Terraform code; both must pass without findings above informational level
- Integration tests (testcontainers‑go) execute in CI and succeed
- Branch protection prevents direct pushes to `main` and `develop`
- Secrets Store CSI Driver and cert‑manager remain functional after automated deployments
**Plans**: 6 plans

1. Create GitHub Actions workflow with OIDC auth for both clouds (covers REQ-IDs: CICD-01)
2. Add Trivy container scan step (covers REQ-IDs: CICD-02)
3. Add tfsec Terraform scan step (covers REQ-IDs: CICD-03)
4. Include integration test job using testcontainers‑go (covers REQ-IDs: CICD-04)
5. Enforce branch protection rules (covers REQ-IDs: CICD-05)
6. Validate secret management post‑deploy (covers REQ-IDs: SECR-01, SECR-02, SECR-03)

## Progress Table

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1 - Architecture & Core App | 1/1 | Complete | 2026-05-07 |
| 2 - Sidecar Proxy | 5/5 | Complete | 2026-05-07 |
| 3 - Terraform Infrastructure | 7/7 | Complete | 2026-05-07 |
| 4 - Observability | 0/5 | Not started | - |
| 5 - CI/CD & Security | 0/6 | Not started | - |

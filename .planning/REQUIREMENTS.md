# Requirements: Multicloud PoC (AWS/Azure) with Go Clean Architecture

**Defined:** 2026-05-06
**Core Value:** Cloud-agnostic Go application using Clean Architecture and sidecar pattern for multicloud capability

## v1 Requirements

### Architecture & Core App
- [x] **ARCH-01**: Main Go application compiles and runs with zero AWS/Azure SDK imports
- [x] **ARCH-02**: Clean Architecture layers implemented (domain, usecase, infrastructure)
- [x] **ARCH-03**: Dependency injection used throughout, no `init()` or global state
- [x] **ARCH-04**: All I/O functions receive `context.Context` as first argument
- [x] **ARCH-05**: Graceful shutdown handles SIGINT/SIGTERM in both app and sidecar

### Sidecar Proxy
- [ ] **SIDE-01**: Sidecar exposes unified gRPC/HTTP API on localhost:50051
- [ ] **SIDE-02**: Sidecar implements AWS DynamoDB operations (Get, Put, Query, Delete)
- [ ] **SIDE-03**: Sidecar implements Azure CosmosDB operations (Get, Create, Query, Delete)
- [ ] **SIDE-04**: Sidecar uses AWS SDK v2 (aws-sdk-go-v2) for DynamoDB
- [ ] **SIDE-05**: Sidecar uses Azure SDK for Go (azcosmos) for CosmosDB
- [ ] **SIDE-06**: Sidecar authenticates via IRSA (AWS) and Workload Identity (Azure)

### Terraform Infrastructure
- [ ] **TERR-01**: AWS VPC module with private subnets, no public exposure
- [ ] **TERR-02**: Azure VNet module with private subnets, no public exposure
- [ ] **TERR-03**: AWS EKS cluster with IRSA OIDC provider enabled
- [ ] **TERR-04**: Azure AKS cluster with Workload Identity enabled
- [ ] **TERR-05**: AWS DynamoDB table (On-Demand) with encryption
- [ ] **TERR-06**: Azure CosmosDB account (SQL API) with appropriate consistency
- [ ] **TERR-07**: Terraform state backends: S3+DynamoDB (AWS), Blob+Lease (Azure)
- [ ] **TERR-08**: Symmetric Terraform module structure across both clouds

### Observability
- [ ] **OBSV-01**: OpenTelemetry SDK integrated in both app and sidecar
- [ ] **OBSV-02**: Traces propagate from app through sidecar to cloud services
- [ ] **OBSV-03**: Metrics exported to AWS X-Ray and Azure Monitor
- [ ] **OBSV-04**: Structured logging with `log/slog` and trace_id correlation
- [ ] **OBSV-05**: OTel Collector deployed as DaemonSet for dual-cloud export

### CI/CD & Security
- [ ] **CICD-01**: GitHub Actions pipeline with OIDC authentication (no static secrets)
- [ ] **CICD-02**: Trivy security scans in pipeline
- [ ] **CICD-03**: tfsec Terraform security scans in pipeline
- [ ] **CICD-04**: Integration tests run in pipeline using testcontainers-go
- [ ] **CICD-05**: No direct commits to `main` or `develop` branches
- [ ] **SECR-01**: Secrets Store CSI Driver configured for both clouds
- [ ] **SECR-02**: cert-manager installed with ClusterIssuers for TLS
- [ ] **SECR-03**: All resources deployed in private subnets

### Testing
- [x] **TEST-01**: Unit tests cover domain and usecase layers (>80% coverage)
- [ ] **TEST-02**: Integration tests validate sidecar with LocalStack (AWS) and Cosmos Emulator (Azure)
- [x] **TEST-03**: TDD Red-Green-Refactor cycle followed for all features
- [ ] **TEST-04**: Testcontainers-go spins up emulators for integration tests
- [ ] **TEST-05**: Performance targets met: <=5ms sidecar latency (p95), >=1000 rps reads

## v2 Requirements
*(Deferred to future release)*

## Traceability
| Requirement | Phase | Status |
|-------------|-------|--------|
| ARCH-01 | Phase 1 | Complete |
| ARCH-02 | Phase 1 | Complete |
| ARCH-03 | Phase 1 | Complete |
| ARCH-04 | Phase 1 | Complete |
| ARCH-05 | Phase 1 | Complete |
| SIDE-01 | Phase 2 | Pending |
| SIDE-02 | Phase 2 | Pending |
| SIDE-03 | Phase 2 | Pending |
| SIDE-04 | Phase 2 | Pending |
| SIDE-05 | Phase 2 | Pending |
| SIDE-06 | Phase 2 | Pending |
| TERR-01 | Phase 3 | Pending |
| TERR-02 | Phase 3 | Pending |
| TERR-03 | Phase 3 | Pending |
| TERR-04 | Phase 3 | Pending |
| TERR-05 | Phase 3 | Pending |
| TERR-06 | Phase 3 | Pending |
| TERR-07 | Phase 3 | Pending |
| TERR-08 | Phase 3 | Pending |
| OBSV-01 | Phase 4 | Pending |
| OBSV-02 | Phase 4 | Pending |
| OBSV-03 | Phase 4 | Pending |
| OBSV-04 | Phase 4 | Pending |
| OBSV-05 | Phase 4 | Pending |
| CICD-01 | Phase 5 | Pending |
| CICD-02 | Phase 5 | Pending |
| CICD-03 | Phase 5 | Pending |
| CICD-04 | Phase 5 | Pending |
| CICD-05 | Phase 5 | Pending |
| SECR-01 | Phase 5 | Pending |
| SECR-02 | Phase 5 | Pending |
| SECR-03 | Phase 5 | Pending |
| TEST-01 | Phase 1 | Complete |
| TEST-02 | Phase 2 | Pending |
| TEST-03 | Phase 1 | Complete |
| TEST-04 | Phase 2 | Pending |
| TEST-05 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 37 total
- Mapped to phases: 37
- Unmapped: 0 ✓

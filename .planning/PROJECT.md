# PROJECT.md

## Project Overview

**Project Name**: Multicloud PoC (AWS/Azure) with Go Clean Architecture

**Description**: A Proof of Concept implementing a cloud-agnostic Go application using Clean Architecture principles. The system uses a sidecar pattern to achieve multicloud capability, with the main application remaining completely agnostic to cloud providers (AWS/Azure). All cloud-specific logic (DynamoDB, CosmosDB, SDKs) is isolated in a sidecar component that communicates with the main app via localhost.

**Problem Statement**: Organizations need applications that can operate across multiple cloud providers without code changes. Current approaches often leak cloud SDK dependencies into core business logic, making it impossible to switch providers or operate in a multicloud environment.

**Solution**: Implement a sidecar proxy pattern where:

- Main Go application (`cmd/app`) contains pure business logic with no cloud SDKs
- Sidecar (`cmd/sidecar`) holds AWS/Azure SDKs and exposes a unified local API
- Communication via gRPC/HTTP on localhost:50051
- Terraform manages symmetric infrastructure across both clouds

## Architecture Principles

### Clean Architecture Layers

1. **Domain Layer** (`internal/domain`): Pure Go entities and interfaces, zero external dependencies
2. **Use Case Layer** (`internal/usecase`): Business logic orchestrating domain interfaces
3. **Infrastructure Layer** (`internal/infrastructure`): Implementations of domain interfaces
4. **Sidecar** (`cmd/sidecar`): Cloud SDK implementations (AWS DynamoDB, Azure CosmosDB)

### Key Design Decisions

- **Zero SDK Leakage**: Main app NEVER imports AWS/Azure SDKs
- **Dependency Injection**: All components wired via constructors, no `init()` or globals
- **TDD Mandatory**: Red-Green-Refactor cycle for all features
- **Context Propagation**: All I/O functions receive `context.Context` as first argument
- **Observability**: OpenTelemetry for traces, metrics, and logs with `trace_id` correlation

## Cloud Providers

- **AWS**: DynamoDB, IRSA for identity, S3/DynamoDB for Terraform state
- **Azure**: CosmosDB, Workload Identity, Blob Storage/Lease for Terraform state

## Identity & Security

- **Zero Static Secrets**: No Access Keys or static credentials
- **OIDC Federation**: GitHub Actions authenticates via OIDC to both clouds
- **IRSA (AWS)**: IAM Roles for Service Accounts in Kubernetes
- **Workload Identity (Azure)**: Federated identity for GKE/AKS workloads

## Infrastructure

- **Terraform Symmetric Modules**: Reusable, standardized modules for both clouds
- **Landing Zones**: Private subnets, Private Links/Endpoints, no public exposure
- **State Management**: S3+DynamoDB (AWS), Blob+Lease (Azure)
- **Secrets**: Secrets Store CSI Driver with cert-manager for TLS

## CI/CD

- **Git Flow**: No direct commits to `main` or `develop`
- **GitHub Actions**: Trivy/tfsec scans, integration tests, OIDC authentication
- **Pipeline Rigor**: Security scans and integration tests before any deployment

## Requirements

### Validated

- [x] Implement Go application with Clean Architecture layers — Validated in Phase 1: Architecture & Core App

### Active

- [x] Implement Go application with Clean Architecture layers *(Phase 1 complete)*
- [ ] Create sidecar proxy for cloud-agnostic communication
- [ ] Implement Terraform modules for AWS (DynamoDB, networking, identity)
- [ ] Implement Terraform modules for Azure (CosmosDB, networking, identity)
- [ ] Set up OpenTelemetry observability (traces, metrics, logs)
- [ ] Configure GitHub Actions with OIDC authentication
- [ ] Implement TDD lifecycle with testcontainers-go for integration tests
- [ ] Set up Landing Zones with private networking
- [ ] Configure Secrets Store CSI Driver and cert-manager

### Out of Scope

- [Multi-region deployment] — Focus on single region per cloud for PoC
- [Production migration tooling] — PoC only, no migration utilities
- [Cost optimization] — Not a priority for initial PoC
- [GUI/Dashboard] — CLI and API only

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Sidecar Pattern | Decouple cloud SDKs from core logic | Main app portable across clouds |
| Clean Architecture | Enforce strict layer separation | Testable, maintainable codebase |
| TDD Mandatory | Ensure quality from start | High coverage, fewer regressions |
| gRPC for App-Sidecar | High performance, typed contracts | Fast, reliable communication |
| Symmetric Terraform | Same module structure for both clouds | Consistent, comparable infrastructure |

## Success Criteria

1. Main Go app compiles and runs WITHOUT any cloud SDK imports
2. Sidecar successfully proxies requests to both AWS DynamoDB and Azure CosmosDB
3. All layers have >80% test coverage (as per `.claude/hardness/test-coverage-thresholds.md`)
4. Terraform modules deploy identical architecture to both clouds
5. OpenTelemetry traces propagate from app through sidecar to cloud services
6. GitHub Actions pipeline passes with OIDC authentication (no static secrets)
7. All resources deployed in private subnets with proper security controls

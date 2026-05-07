# Multicloud PoC Technical Research

**Project:** Multicloud PoC (AWS/Azure) with Go Clean Architecture
**Date:** 2026-05-06

## Cloud Services

| Domain | AWS Service | Azure Service | Reason for Selection |
|--------|-------------|---------------|----------------------|
| NoSQL DB | DynamoDB (On-Demand) | Cosmos DB (SQL API) | Both provide fully managed, schema-less storage with low-latency read/write. On-Demand avoids capacity planning for a PoC while Cosmos offers multi-model flexibility. |
| Container Orchestration | Amazon EKS | Azure AKS | Native managed Kubernetes clusters to host the sidecar pattern. Both integrate with respective IAM and networking primitives. |
| Identity / Federation | IAM Roles for Service Accounts (IRSA) + OIDC Provider | Azure Workload Identity (User-Assigned Managed Identity) | Enables zero-trust, least-privilege access for sidecar pods without embedding long-lived credentials. |
| Secrets Management | AWS Secrets Manager (via Secrets Store CSI Driver) | Azure Key Vault (via Secrets Store CSI Driver) | Centralised secret storage with automatic rotation; CSI driver abstracts the provider for Kubernetes pods. |
| Observability | OpenTelemetry Collector -> AWS Managed Prometheus & X-Ray | OpenTelemetry Collector -> Azure Monitor (Application Insights) & Log Analytics | Uniform OTel telemetry enables a single view across clouds; collector can export to both back-ends simultaneously. |
| TLS & Certificates | Cert-Manager + AWS Private CA / Let's Encrypt (DNS-01 via Route53) | Cert-Manager + Azure Key Vault (DNS-01 via Azure DNS) | Automated certificate lifecycle; mTLS not required for localhost sidecar-app traffic but TLS-1.2+ mandatory for cloud SDK calls. |
| CI/CD Artifact Registry | Amazon ECR (signed with Cosign) | Azure Container Registry (signed with Cosign) | Secure, signed container images for sidecar and main app. |
| Infrastructure State | S3 bucket + DynamoDB lock table | Azure Blob Storage + azurerm lock (Blob lease) | Symmetric back-ends per provider respecting Terraform Standards spec. |

## SDKs

| Layer | AWS SDK | Azure SDK |
|-------|---------|-----------|
| Go sidecar | `github.com/aws/aws-sdk-go-v2` (v2) - DynamoDB client, STS for token refresh, IAM for role assumption | `github.com/Azure/azure-sdk-for-go/sdk/azidentity` + `azcosmos` - Cosmos DB client, Managed Identity handling |
| Common | OpenTelemetry Go SDK (`go.opentelemetry.io/otel`), `log/slog` for structured logs, `google.golang.org/grpc` for gRPC sidecar API | Same common libs; avoid provider-specific imports in core app. |

**Why these versions:** Both SDKs are the latest major releases (v2 for AWS, v0.30+ for Azure) and support context propagation, retries, and exponential back-off out-of-the-box - required by the *Reliability* pillar in the AWS Well-Architected spec and the *Security* pillar in Azure.

## Terraform Modules

| Module | Purpose | Key Outputs (per spec) |
|--------|---------|------------------------|
| `aws/vpc` | Creates Hub-and-Spoke VPC, NAT, Gateway Endpoints for DynamoDB, S3 | `vpc_id`, `public_subnet_ids`, `private_subnet_ids` |
| `azure/vnet` | Hub-and-Spoke VNet with Azure Firewall and Private Link for Cosmos DB | `vnet_id`, `subnet_ids` |
| `aws/eks` | EKS cluster with IRSA OIDC provider, node groups, IAM policies for sidecar | `cluster_name`, `oidc_provider_arn`, `node_role_arn` |
| `azure/aks` | AKS cluster with workload identity enablement | `cluster_name`, `identity_profile` |
| `aws/dynamodb` | DynamoDB table (On-Demand) with TTL, encryption, IAM policy output | `table_name`, `table_arn`, `iam_policy_arn` |
| `azure/cosmosdb` | Cosmos DB account (SQL API) with consistency level and RBAC role output | `account_endpoint`, `account_id`, `rbac_role_id` |
| `observability/otel-collector` | Deploys OTel Collector as DaemonSet or sidecar, config for dual export | `collector_service_name` |
| `certs/cert-manager` | Installs cert-manager, creates ClusterIssuers for AWS Private CA & Azure Key Vault | N/A |
| `secrets/csi-driver` | Deploys Secrets Store CSI Driver and SecretProviderClass resources for both clouds | N/A |

**Design notes:** Modules follow the *Terraform Standards* spec - each resource is encapsulated, version-pinned providers, and standardized tags (`project_name`, `managed_by`, `environment`). Outputs defined match the *Cloud-Resource-Mapping* spec requirements for downstream wiring.

## Trade-offs & Architectural Considerations

| Trade-off | Impact | Mitigation / Recommendation |
|----------|--------|------------------------------|
| **Additional network hop (App -> Sidecar -> Cloud)** | Adds <=5ms latency target (p95) and extra CPU/memory for serialization. | Keep payloads minimal (Protobuf), pre-warm connection pools, and enforce 100% OTel sampling in PoC to detect bottlenecks. |
| **Duplicate SDK footprints** (AWS + Azure) | Increases container image size (~30MB). | Build multi-stage Dockerfile; sidecar only includes required SDKs per target via build-time `GOOS=linux GOARCH=amd64 go build -tags aws,azure`. |
| **State management divergence** (S3 vs Blob) | Potential drift in Terraform state handling. | Use Terraform workspaces per provider and an abstracted backend script that syncs state files to a central S3 bucket for audit. |
| **Identity rotation latency** | Token refresh may pause sidecar requests. | Implement background watcher with `inotify` (as per Secrets Management spec) to reload tokens without pod restart. |
| **Observability duplication** | Exporting to two back-ends can double data volume. | Enable sampling control per environment; for PoC use 100% to validate, switch to 10% for production. |
| **Cost of dual managed services** | Running DynamoDB and CosmosDB incurs double spend. | PoC uses On-Demand/DynamoDB and low-tier Cosmos DB; cost monitoring via Terraform cost-estimate and tag-based alerts. |

## Recommendations

1. **Adopt the sidecar abstraction** as the single source of truth for cloud interactions. It satisfies the *Sidecar-Abstraction* spec and enforces zero-leakage of SDKs into core business code.
2. **Standardize on OpenTelemetry** for tracing, metrics, and logs. Export both to AWS X-Ray and Azure Monitor via a unified collector - this aligns with the *Observability-Stack* and *Observability-Rigor* hardness specifications.
3. **Leverage provider-native IAM federation** (IRSA & Azure Workload Identity) to meet the *Security-Compliance* and *Cross-Cloud-Resilience* hardness requirements while avoiding secret sprawl.
4. **Implement Terraform modules per the *Terraform-Standards* spec** - keep modules isolated, versioned, and tag-rich to simplify governance and future extension to additional clouds.
5. **Validate performance targets early** - use `testcontainers-go` with LocalStack and Cosmos Emulator to benchmark sidecar latency (<5ms) and throughput (>=1000 rps reads, 500 rps writes) as defined in the *Performance-Targets* hardness file.
6. **Integrate chaos engineering** (network partitions, service outages) into CI pipeline per the *Resilience* hardness spec to ensure circuit-breaker behavior and automated rollback.

---

## Sources

- .claude/specs/technical/cloud-resource-mapping.md
- .claude/specs/infrastructure/sidecar-abstraction.md
- .claude/specs/infrastructure/app-go-clean-architecture.md
- .claude/specs/technical/terraform-standards.md
- .claude/specs/infrastructure/observability-stack.md
- .claude/specs/infrastructure/observability-pipelines.md
- .claude/specs/technical/go-observability-otel.md
- .claude/specs/technical/golang-implementation-standards.md
- .claude/specs/technical/iam-federation.md
- .claude/specs/well-architected/aws-pillars.md
- .claude/specs/well-architected/azure-pillars.md
- .claude/hardness/performance-targets.md
- .claude/hardness/cross-cloud-resilience.md
- .claude/hardness/observability-rigor.md
- .claude/hardness/security-compliance.md
- .claude/hardness/test-coverage-thresholds.md
- .claude/hardness/cicd-rigor.md

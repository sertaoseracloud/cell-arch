# Phase 3: Terraform Infrastructure - Context

**Gathered:** 2026-05-06
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase delivers symmetric Terraform modules that provision private networking, managed Kubernetes clusters, managed databases, and secret management tooling across AWS and Azure. Three environments (dev / hom / prod) are supported via per-environment `.tfvars` files and isolated state backends. The main application is NOT deployed here — this phase is infrastructure-only.
</domain>

<decisions>
## Implementation Decisions

### Environment Strategy

- **D-17:** Three environments: `dev`, `hom` (homologation/staging), `prod`.
- **D-18:** Environment-specific values in `.tfvars` files — one per environment per cloud (e.g., `dev.tfvars`, `hom.tfvars`, `prod.tfvars`).
- **D-19:** Separate state backends per environment — each of dev / hom / prod gets its own S3 bucket + DynamoDB table (AWS) and its own Azure Blob container + Lease lock (Azure). Full blast-radius isolation between environments.

### State Backend Bootstrapping

- **D-20:** Bootstrap module at `infra/bootstrap/` using local Terraform state (not remote). Creates the S3 bucket, DynamoDB lock table, Azure Blob container, and Lease lock for each environment. Run once per environment before the main modules. After bootstrap, main modules use remote backends.
- **D-21:** Bootstrap is per-environment: one bootstrap invocation per env (dev, hom, prod). Naming convention: `{project}-tfstate-{env}` for both AWS and Azure.

### Node Sizing (EKS / AKS)

- **D-22:** Instance type: `t3.medium` (AWS EKS) and `Standard_D2s_v3` (Azure AKS) — 2 vCPU, 4 GB RAM. Sufficient for app + sidecar + system pods (CoreDNS, cert-manager, Secrets Store CSI Driver).
- **D-23:** Node count: 2 nodes minimum per cluster per environment. Required for cert-manager and Secrets Store CSI Driver HA.

### CosmosDB Configuration

- **D-24:** Consistency level: **Session** — guarantees read-your-writes within a session; matches DynamoDB's per-session semantics. No extra cost vs Eventual.
- **D-25:** Single region, no geo-redundancy (multi-region writes disabled). PoC scope — geo-redundant writes add ~2x cost and conflict resolution complexity not needed here.

### Module Structure (from terraform-standards.md — locked)

- Encapsulate EKS, AKS, DynamoDB, CosmosDB in reusable modules under `infra/modules/`.
- Live environments under `infra/live/{aws,azure}/{dev,hom,prod}/`.
- Each cloud directory has its own `providers.tf` with pinned provider versions.
- Mandatory module outputs: `database_endpoint`, `resource_id`, `iam_policy_arn` (AWS) / equivalent policy object ID (Azure).
- Mandatory resource tags: `project_name`, `managed_by: terraform`, `environment: {dev|hom|prod}`.

### Networking (from landing-zones.md — locked)

- AWS: Hub-Spoke VPC topology. Hub centralizes NAT Gateways + traffic inspection. Spoke hosts EKS. DynamoDB accessed via VPC Gateway Endpoint (no public traffic).
- Azure: Hub-Spoke VNet topology. Hub contains Azure Firewall. Spoke hosts AKS. CosmosDB accessed via Azure Private Link (no public IP on database).
- All compute resources in private subnets. No public subnets.

### Claude's Discretion

- **CIDR Ranges** (discussed 2026-05-07):
  - Hub VPC: `10.0.0.0/16` (AWS Hub)
  - AWS Spoke VPC: `10.1.0.0/16` (EKS private subnets)
  - Azure Hub VNet: `10.2.0.0/16`
  - Azure Spoke VNet: `10.3.0.0/16`
  - Each spoke subnet: `/24` (254 IPs per subnet, enough for 2-node clusters + system pods)
  - No overlap between AWS and Azure ranges.
- **Kubernetes version** (discussed 2026-05-07): Pin EKS and AKS to `1.29` (latest stable supported by both clouds as of 2026-05).
- **Terraform and provider versions** (discussed 2026-05-07):
  - Terraform: `1.7.5`
  - AWS Provider: `5.48.0` (supports IRSA, DynamoDB, VPC endpoints)
  - Azure Provider: `3.101.0` (supports AKS, CosmosDB, Private Link)
  - Azure Go SDK (`azcosmos`, `azidentity`): `v1.x.x` (already pinned in go.mod)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Architecture
- `.planning/ROADMAP.md` — Phase 3 goals, success criteria, 7-plan breakdown
- `.planning/PROJECT.md` — Symmetric Terraform principle, zero static secrets, landing zone requirements
- `.planning/REQUIREMENTS.md` — TERR-01 through TERR-08, SECR-01, SECR-02, SECR-03

### Technical Standards
- `.claude/specs/technical/terraform-standards.md` — Module structure, state backend symmetry, required outputs, mandatory tagging (`project_name`, `managed_by: terraform`, `environment: poc`)
- `.claude/specs/infrastructure/landing-zones.md` — Hub-Spoke networking topology, Gateway Endpoints (AWS), Private Link (Azure), SCPs / Azure Policy
- `.claude/specs/infrastructure/secrets-management.md` — Secrets Store CSI Driver configuration
- `.claude/specs/infrastructure/certificates-and-tls.md` — cert-manager ClusterIssuers for mTLS

### Prior Phase Context
- `.planning/phases/01-architecture-core-app/01-CONTEXT.md` — Established stack decisions (D-01 through D-08)
- `.planning/phases/02-sidecar-proxy/01-CONTEXT.md` — IRSA + Workload Identity (D-10), mTLS (D-15), per-request cloud selector (D-12)

### Compliance & Hardness
- `.claude/hardness/security-rules.md` — No static credentials, private subnets required
- `.claude/hardness/performance-budgets.md` — Sidecar latency ≤5 ms p95 (infrastructure must not bottleneck)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `infra/` — directory exists but is empty; all Terraform code is greenfield for this phase.

### Established Patterns
- Symmetric structure: every AWS resource has an Azure equivalent with identical interface (from PROJECT.md + terraform-standards.md).
- Manual constructors / DI pattern (D-02) applies to Go code in app/sidecar — not directly applicable to Terraform, but the symmetry principle mirrors it.

### Integration Points
- `cmd/sidecar/main.go` — reads `AWS_REGION`, `DYNAMODB_TABLE`, `AZURE_COSMOS_ENDPOINT`, `COSMOS_DATABASE`, `COSMOS_CONTAINER` from env. Terraform outputs must provide these values.
- EKS/AKS clusters will host both `cmd/app` and `cmd/sidecar` as Kubernetes Deployments (Phase 4+).
- cert-manager ClusterIssuer will issue the mTLS certificates that `cmd/sidecar/main.go` loads via `TLS_CERT`/`TLS_KEY` env vars.
- IRSA role ARN and Workload Identity client ID are outputs that feed the Kubernetes ServiceAccount annotations (wired in Phase 5).

</code_context>

<specifics>
## Specific Ideas

### Directory Layout

```
infra/
├── bootstrap/          # Local-state bootstrap module (creates state backends)
│   ├── aws/
│   └── azure/
├── modules/            # Reusable modules
│   ├── aws-vpc/
│   ├── aws-eks/
│   ├── aws-dynamodb/
│   ├── azure-vnet/
│   ├── azure-aks/
│   └── azure-cosmosdb/
└── live/               # Environment-specific roots
    ├── aws/
    │   ├── dev/        # dev.tfvars + main.tf wiring modules
    │   ├── hom/
    │   └── prod/
    └── azure/
        ├── dev/
        ├── hom/
        └── prod/
```

### Required Terraform Outputs per Database Module
- `database_endpoint` — used by sidecar env vars
- `resource_id` — used for IAM policy attachment
- `iam_policy_arn` (AWS) / `policy_object_id` (Azure) — attached to IRSA role / Workload Identity

### State Backend Naming
- AWS S3 bucket: `{project_name}-tfstate-aws-{env}` (e.g., `cell-arch-tfstate-aws-dev`)
- AWS DynamoDB lock table: `{project_name}-tflock-aws-{env}`
- Azure Blob container: `{project_name}-tfstate-azure-{env}`

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within Phase 3 scope.

</deferred>

---
*Phase: 03-terraform-infrastructure*
*Context gathered: 2026-05-06*

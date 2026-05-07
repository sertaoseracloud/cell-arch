# Phase 3: Terraform Infrastructure - Research

**Researched:** 2026-05-06
**Domain:** Terraform IaC — AWS (EKS, DynamoDB, VPC) + Azure (AKS, CosmosDB, VNet) with Helm chart installs
**Confidence:** MEDIUM-HIGH (provider versions verified; architecture patterns cross-referenced)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-17:** Three environments: `dev`, `hom` (homologation/staging), `prod`.
- **D-18:** Environment-specific values in `.tfvars` files — one per environment per cloud.
- **D-19:** Separate state backends per environment — each env gets its own S3 bucket + DynamoDB lock table (AWS) and its own Azure Blob container + Lease lock (Azure). Full blast-radius isolation.
- **D-20:** Bootstrap module at `infra/bootstrap/` using local Terraform state. Creates the S3 bucket, DynamoDB lock table, Azure Blob container, and Lease lock for each environment. Runs once per environment before main modules. After bootstrap, main modules use remote backends.
- **D-21:** Bootstrap is per-environment. Naming: `{project}-tfstate-{env}` for both clouds.
- **D-22:** Instance type: `t3.medium` (AWS EKS) and `Standard_D2s_v3` (Azure AKS).
- **D-23:** Node count: 2 nodes minimum per cluster per environment.
- **D-24:** CosmosDB consistency level: **Session**.
- **D-25:** Single region, no geo-redundancy.
- Module structure from `terraform-standards.md`: modules under `infra/modules/`, live envs under `infra/live/{aws,azure}/{dev,hom,prod}/`.
- Mandatory outputs: `database_endpoint`, `resource_id`, `iam_policy_arn` (AWS) / equivalent (Azure).
- Mandatory tags: `project_name`, `managed_by: terraform`, `environment: {env}`.
- AWS Hub-Spoke VPC: Hub centralizes NAT + inspection; Spoke hosts EKS; DynamoDB via Gateway Endpoint.
- Azure Hub-Spoke VNet: Hub contains Azure Firewall; Spoke hosts AKS; CosmosDB via Azure Private Link.
- All compute in private subnets. No public subnets.
- IRSA (AWS) + Workload Identity (Azure) — no static credentials.
- cert-manager + Secrets Store CSI Driver in both clusters.

### Claude's Discretion

- Specific CIDR ranges for VPC/VNet and subnets.
- Kubernetes version — pin to latest stable EKS/AKS-supported at time of planning.
- Terraform and provider version pins.

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within Phase 3 scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TERR-01 | AWS VPC module with private subnets, no public exposure | Hub-Spoke VPC + Gateway Endpoint pattern confirmed |
| TERR-02 | Azure VNet module with private subnets, no public exposure | Hub-Spoke VNet Peering + Private Link confirmed |
| TERR-03 | AWS EKS cluster with IRSA OIDC provider enabled | aws_iam_openid_connect_provider + tls_certificate datasource pattern |
| TERR-04 | Azure AKS cluster with Workload Identity enabled | oidc_issuer_enabled + workload_identity_enabled + azurerm_federated_identity_credential |
| TERR-05 | AWS DynamoDB table (On-Demand) with encryption | billing_mode=PAY_PER_REQUEST + server_side_encryption block |
| TERR-06 | Azure CosmosDB account (SQL API) with Session consistency | azurerm_cosmosdb_account + consistency_policy block |
| TERR-07 | Terraform state backends: S3+use_lockfile (AWS), Blob+Lease (Azure) | S3 native locking confirmed; azurerm backend uses Blob Lease by default |
| TERR-08 | Symmetric Terraform module structure across both clouds | Directory layout and output/tagging standards confirmed |
| SECR-01 | Secrets Store CSI Driver configured for both clouds | Helm chart v1.6.0 (driver) + cloud provider charts confirmed |
| SECR-02 | cert-manager installed with ClusterIssuers for TLS | cert-manager v1.20.2 Helm chart confirmed, crds.enabled=true |
| SECR-03 | All resources deployed in private subnets | Private subnet design enforced by Hub-Spoke + no public subnets |
</phase_requirements>

---

## Summary

This phase provisions symmetric AWS and Azure infrastructure for a multicloud Go PoC: private networking (Hub-Spoke VPC/VNet), managed Kubernetes clusters (EKS/AKS), managed databases (DynamoDB/CosmosDB), and cluster tooling (cert-manager, Secrets Store CSI Driver). Everything must be expressed as reusable Terraform modules with three independent environment copies (dev/hom/prod), isolated state backends, zero public database exposure, and workload-identity-based authentication (IRSA on AWS, Federated Workload Identity on Azure).

The Terraform ecosystem has moved forward significantly. The AWS provider is now on v6 (released June 2025) with a new multi-region attribute model and deprecated S3+DynamoDB lock in favor of S3 native locking (`use_lockfile = true`). The Azure provider is at v4.x. Terraform CLI is at v1.15 but v1.9+ is sufficient for all features needed here. cert-manager dropped the old `installCRDs` flag in v1.15+ in favor of `crds.enabled=true`. Secrets Store CSI Driver v1.6.0 switched to the RequiresRepublish kubelet mechanism for secret rotation.

**Primary recommendation:** Use raw `aws_eks_cluster` + `aws_eks_node_group` + `aws_eks_addon` resources (not the community module) to maintain full visibility and auditability; use `azurerm_kubernetes_cluster` directly. Deploy cert-manager and Secrets Store CSI Driver via `helm_release` with exec-based Kubernetes provider authentication to avoid the 15-minute token expiry bug. Use S3 native locking (`use_lockfile = true`) and drop the DynamoDB lock table in bootstrap for AWS.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Private networking (subnets, routing) | Terraform modules (aws-vpc / azure-vnet) | — | Pure IaC, no runtime tier |
| DynamoDB Gateway Endpoint | Terraform aws-vpc module | — | Route-table resource, owned by VPC module |
| CosmosDB Private Link + DNS | Terraform azure-cosmosdb module | azure-vnet module (subnet) | Private endpoint depends on subnet from vnet module |
| EKS cluster + managed node group | Terraform aws-eks module | — | aws_eks_cluster + aws_eks_node_group |
| IRSA OIDC provider + IAM trust policy | Terraform aws-eks module | — | Outputs fed to sidecar ServiceAccount annotations |
| AKS cluster + system node pool | Terraform azure-aks module | — | azurerm_kubernetes_cluster |
| AKS Workload Identity + federated cred | Terraform azure-aks module | — | azurerm_federated_identity_credential |
| DynamoDB table | Terraform aws-dynamodb module | — | PAY_PER_REQUEST + SSE |
| CosmosDB account + database + container | Terraform azure-cosmosdb module | — | SQL API with Session consistency |
| cert-manager Helm chart | Terraform aws-eks / azure-aks live root | — | helm_release in cluster root after cluster ready |
| Secrets Store CSI Driver + provider | Terraform aws-eks / azure-aks live root | — | helm_release after cert-manager |
| State backend (S3 / Blob) | Terraform bootstrap modules | — | Run once per env, local state |

---

## Standard Stack

### Core Providers

| Provider | Version | Purpose | Why Standard |
|----------|---------|---------|--------------|
| hashicorp/aws | `~> 6.0` | All AWS resources | Latest GA (6.44 as of May 2026); multi-region attribute model, S3 native lock |
| hashicorp/azurerm | `~> 4.0` | All Azure resources | Latest GA (4.71 as of May 2026); full AKS Workload Identity support |
| hashicorp/kubernetes | `~> 3.1` | Kubernetes resources post-cluster | v3 uses plugin protocol v6; v3.1.0 latest |
| hashicorp/helm | `~> 3.1` | Helm releases (cert-manager, CSI) | v3.1.1 latest; plugin protocol v6 |
| hashicorp/tls | `~> 4.0` | Fetch EKS OIDC thumbprint via TLS data source | Required for aws_iam_openid_connect_provider |

**Minimum Terraform CLI:** `>= 1.10` — required for S3 native locking (`use_lockfile`); v1.9.8 is installed locally and lacks this feature. Recommend pinning `>= 1.10` in `required_version`. [VERIFIED: developer.hashicorp.com/terraform/language/backend/s3]

**Version block template:**
```hcl
terraform {
  required_version = ">= 1.10"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 3.1"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.1"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}
```
[VERIFIED: registry.terraform.io, releases.hashicorp.com — versions confirmed May 2026]

### Helm Charts

| Chart | Repo / OCI | Version | Purpose |
|-------|-----------|---------|---------|
| cert-manager | `oci://quay.io/jetstack/charts/cert-manager` | `v1.20.2` | TLS certificate issuance |
| secrets-store-csi-driver | `https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts` | `1.6.0` | CSI secrets volume driver |
| secrets-store-csi-driver-provider-aws | `https://aws.github.io/secrets-store-csi-driver-provider-aws` | latest (floats) | AWS Secrets Manager provider |
| csi-secrets-store-provider-azure | `https://azure.github.io/secrets-store-csi-driver-provider-azure/charts` | `1.8.1` | Azure Key Vault provider |

[VERIFIED: cert-manager.io/docs/installation/helm — v1.20.2 confirmed; github.com/kubernetes-sigs/secrets-store-csi-driver/releases — v1.6.0 Apr 2026; artifacthub.io — azure provider 1.8.1]

### Installation (bootstrap one-time)

```bash
# Terraform CLI upgrade required (local is v1.9.8, need >= 1.10)
# Install tfenv or download from releases.hashicorp.com/terraform/1.15.1/

terraform -chdir=infra/bootstrap/aws init && terraform -chdir=infra/bootstrap/aws apply
terraform -chdir=infra/bootstrap/azure init && terraform -chdir=infra/bootstrap/azure apply
```

---

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────── AWS live/{dev,hom,prod} ─────────────────────────┐
│                                                                       │
│  ┌──── Bootstrap (local state) ────┐                                 │
│  │  S3 bucket + use_lockfile=true  │ (run once per env)              │
│  └────────────┬────────────────────┘                                 │
│               │ creates remote backend                               │
│               ▼                                                       │
│  ┌──── aws-vpc module ─────────────────────────────────────┐        │
│  │  Hub VPC (10.0.0.0/16)  ←→  Spoke VPC (10.1.0.0/16)   │        │
│  │  NAT Gateway in Hub          Private subnets only        │        │
│  │  Gateway Endpoint: DynamoDB ──────────────────────────► │        │
│  └─────────────────────────┬───────────────────────────────┘        │
│                             │ vpc_id, subnet_ids                     │
│  ┌──── aws-eks module ──────▼──────────────────────────────┐        │
│  │  aws_eks_cluster  ──→  aws_eks_node_group (t3.medium x2) │       │
│  │  aws_iam_openid_connect_provider (tls_certificate fetch) │        │
│  │  IAM role + trust policy (IRSA)                          │        │
│  │  aws_eks_addon: vpc-cni, coredns, kube-proxy             │        │
│  └─────────────────────────┬───────────────────────────────┘        │
│                             │ cluster_endpoint, oidc_issuer_url      │
│  ┌──── aws-dynamodb module ─▼──────────────────────────────┐        │
│  │  aws_dynamodb_table (PAY_PER_REQUEST + SSE)              │        │
│  │  Outputs: table_name, table_arn, iam_policy_arn          │        │
│  └─────────────────────────────────────────────────────────┘        │
│                                                                       │
│  live root:  helm_release cert-manager  →  helm_release CSI driver  │
│              (depends_on aws_eks_node_group)                         │
└───────────────────────────────────────────────────────────────────────┘

┌─────────────────── Azure live/{dev,hom,prod} ───────────────────────┐
│                                                                       │
│  ┌──── Bootstrap (local state) ────┐                                 │
│  │  Storage Account + Blob + Lease │ (run once per env)              │
│  └────────────┬────────────────────┘                                 │
│               │ creates remote backend                               │
│               ▼                                                       │
│  ┌──── azure-vnet module ──────────────────────────────────┐        │
│  │  Hub VNet (10.10.0.0/16) ←peering→ Spoke VNet           │        │
│  │  (10.11.0.0/16)          Bidirectional peering           │        │
│  │  Private subnets only                                    │        │
│  └─────────────────────────┬───────────────────────────────┘        │
│                             │ vnet_id, subnet_ids                    │
│  ┌──── azure-aks module ────▼──────────────────────────────┐        │
│  │  azurerm_kubernetes_cluster                              │        │
│  │  oidc_issuer_enabled=true, workload_identity_enabled=true│        │
│  │  system pool: Standard_D2s_v3 x2                         │        │
│  │  azurerm_user_assigned_identity (workload identity)      │        │
│  │  azurerm_federated_identity_credential                   │        │
│  └─────────────────────────┬───────────────────────────────┘        │
│                             │ oidc_issuer_url, identity_client_id    │
│  ┌──── azure-cosmosdb module▼──────────────────────────────┐        │
│  │  azurerm_cosmosdb_account (SQL, Session, no-public)      │        │
│  │  azurerm_cosmosdb_sql_database                           │        │
│  │  azurerm_cosmosdb_sql_container                          │        │
│  │  azurerm_private_endpoint (subresource: Sql)             │        │
│  │  azurerm_private_dns_zone (privatelink.documents.azure.com)│      │
│  │  azurerm_private_dns_zone_virtual_network_link           │        │
│  └─────────────────────────────────────────────────────────┘        │
│                                                                       │
│  live root:  helm_release cert-manager  →  helm_release CSI driver  │
│              (depends_on azurerm_kubernetes_cluster)                 │
└───────────────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure

```
infra/
├── bootstrap/
│   ├── aws/
│   │   ├── main.tf          # aws_s3_bucket + aws_s3_bucket_versioning + local backend
│   │   ├── variables.tf     # project_name, environment, region
│   │   └── outputs.tf       # bucket_name, region
│   └── azure/
│       ├── main.tf          # azurerm_resource_group + storage_account + container + local backend
│       ├── variables.tf
│       └── outputs.tf
├── modules/
│   ├── aws-vpc/
│   │   ├── main.tf          # aws_vpc, subnets, route tables, NAT GW, VPC peering, Gateway endpoint
│   │   ├── variables.tf
│   │   └── outputs.tf       # vpc_id, private_subnet_ids, route_table_ids
│   ├── aws-eks/
│   │   ├── main.tf          # aws_eks_cluster, node_group, OIDC provider, IAM role
│   │   ├── variables.tf
│   │   └── outputs.tf       # cluster_endpoint, cluster_ca, oidc_issuer_url, irsa_role_arn, iam_policy_arn
│   ├── aws-dynamodb/
│   │   ├── main.tf          # aws_dynamodb_table
│   │   ├── variables.tf
│   │   └── outputs.tf       # database_endpoint (table name), resource_id (arn), iam_policy_arn
│   ├── azure-vnet/
│   │   ├── main.tf          # azurerm_virtual_network x2 (hub+spoke), subnets, peering x2
│   │   ├── variables.tf
│   │   └── outputs.tf       # spoke_vnet_id, private_subnet_ids
│   ├── azure-aks/
│   │   ├── main.tf          # azurerm_kubernetes_cluster, user_assigned_identity, federated_credential
│   │   ├── variables.tf
│   │   └── outputs.tf       # cluster_endpoint, kube_config, oidc_issuer_url, workload_identity_client_id
│   └── azure-cosmosdb/
│       ├── main.tf          # cosmosdb_account, sql_database, sql_container, private_endpoint, DNS zone
│       ├── variables.tf
│       └── outputs.tf       # database_endpoint, resource_id, policy_object_id
└── live/
    ├── aws/
    │   ├── dev/
    │   │   ├── main.tf      # module calls: aws-vpc, aws-eks, aws-dynamodb
    │   │   ├── providers.tf  # aws provider + kubernetes/helm with exec auth
    │   │   ├── backend.tf    # s3 backend with use_lockfile=true
    │   │   ├── dev.tfvars
    │   │   └── outputs.tf
    │   ├── hom/  (same layout)
    │   └── prod/ (same layout)
    └── azure/
        ├── dev/
        │   ├── main.tf      # module calls: azure-vnet, azure-aks, azure-cosmosdb
        │   ├── providers.tf  # azurerm + kubernetes/helm with exec auth
        │   ├── backend.tf    # azurerm backend
        │   ├── dev.tfvars
        │   └── outputs.tf
        ├── hom/
        └── prod/
```

---

### Pattern 1: EKS IRSA OIDC Provider

**What:** Create OIDC identity provider for EKS in IAM, then configure an IAM role with a trust policy that allows Kubernetes service accounts to assume it via OIDC token exchange.

**When to use:** Required for SIDE-06 — sidecar must authenticate to DynamoDB via IRSA.

```hcl
# Source: oneuptime.com/blog (verified against registry.terraform.io/providers/hashicorp/aws)
data "tls_certificate" "eks" {
  url = aws_eks_cluster.main.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "eks" {
  url             = aws_eks_cluster.main.identity[0].oidc[0].issuer
  client_id_list  = ["sts.amazonaws.com"]
  # For EKS-managed OIDC (backed by AWS-trusted CAs), the thumbprint is
  # not actually used by IAM, but the argument is still required by Terraform.
  thumbprint_list = [data.tls_certificate.eks.certificates[0].sha1_fingerprint]
}

data "aws_iam_policy_document" "irsa_trust" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.eks.arn]
    }
    condition {
      test     = "StringEquals"
      variable = "${replace(aws_iam_openid_connect_provider.eks.url, "https://", "")}:sub"
      values   = ["system:serviceaccount:${var.namespace}:${var.service_account_name}"]
    }
    condition {
      test     = "StringEquals"
      variable = "${replace(aws_iam_openid_connect_provider.eks.url, "https://", "")}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "irsa" {
  name               = "${var.project_name}-irsa-${var.environment}"
  assume_role_policy = data.aws_iam_policy_document.irsa_trust.json
}
```
[VERIFIED: repost.aws, oneuptime.com — cross-referenced against official IRSA pattern]

### Pattern 2: AKS Workload Identity

**What:** Enable OIDC issuer and Workload Identity on AKS, create a User Assigned Managed Identity, and bind it to a Kubernetes ServiceAccount via a Federated Identity Credential.

**When to use:** Required for SIDE-06 — sidecar must authenticate to CosmosDB via Workload Identity.

```hcl
# Source: learn.microsoft.com/en-us/azure/aks/workload-identity-deploy-cluster
resource "azurerm_kubernetes_cluster" "main" {
  name                = "${var.project_name}-aks-${var.environment}"
  location            = var.location
  resource_group_name = azurerm_resource_group.main.name
  dns_prefix          = "${var.project_name}-${var.environment}"

  oidc_issuer_enabled       = true
  workload_identity_enabled = true

  default_node_pool {
    name           = "system"
    node_count     = var.node_count   # min 2
    vm_size        = "Standard_D2s_v3"
    vnet_subnet_id = var.subnet_id
    only_critical_addons_enabled = false
  }

  identity {
    type = "SystemAssigned"
  }

  tags = var.tags
}

resource "azurerm_user_assigned_identity" "sidecar" {
  name                = "${var.project_name}-wi-${var.environment}"
  resource_group_name = azurerm_resource_group.main.name
  location            = var.location
}

resource "azurerm_federated_identity_credential" "sidecar" {
  name                = "${var.project_name}-fic-${var.environment}"
  resource_group_name = azurerm_resource_group.main.name
  parent_id           = azurerm_user_assigned_identity.sidecar.id
  audience            = ["api://AzureADTokenExchange"]
  issuer              = azurerm_kubernetes_cluster.main.oidc_issuer_url
  subject             = "system:serviceaccount:${var.namespace}:${var.service_account_name}"
}
```
[CITED: learn.microsoft.com/en-us/azure/aks/workload-identity-deploy-cluster]

### Pattern 3: S3 Backend with Native Locking

**What:** Use S3's conditional write locking instead of DynamoDB. DynamoDB-based locking is now deprecated in Terraform 1.10+ and scheduled for removal.

**When to use:** All AWS live environments.

```hcl
# Source: developer.hashicorp.com/terraform/language/backend/s3
terraform {
  backend "s3" {
    bucket       = "cell-arch-tfstate-aws-dev"
    key          = "terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true   # S3 native locking via conditional PutObject
    encrypt      = true
  }
}
```

**Required S3 IAM permissions for use_lockfile:**
- `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` on `<bucket>/<key>.tflock`
- Standard `s3:GetObject`, `s3:PutObject` on the state file itself

[VERIFIED: developer.hashicorp.com/terraform/language/backend/s3]

### Pattern 4: Azure Blob Backend

**What:** azurerm backend with Blob container. Locking is automatic via Azure Blob Lease — no extra configuration needed.

```hcl
# Source: developer.hashicorp.com/terraform/language/backend/azurerm
terraform {
  backend "azurerm" {
    resource_group_name  = "cell-arch-tfstate-azure-dev"
    storage_account_name = "cellarchstateazuredev"
    container_name       = "cell-arch-tfstate-azure-dev"
    key                  = "terraform.tfstate"
    # use_azuread_auth = true  # recommended when running in CI with OIDC
  }
}
```
[VERIFIED: developer.hashicorp.com/terraform/language/backend/azurerm]

### Pattern 5: Helm Provider with Exec Auth (EKS)

**What:** Use exec-based token retrieval to avoid 15-minute static token expiry. The `aws eks get-token` command is called on each Terraform invocation.

```hcl
# Source: developer.hashicorp.com/terraform/tutorials/kubernetes/helm-provider
provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_ca)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name, "--region", var.aws_region]
    }
  }
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_ca)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name, "--region", var.aws_region]
  }
}
```
[VERIFIED: hashicorp/terraform-provider-helm GitHub issue #1290 — exec plugin recommended]

### Pattern 6: DynamoDB Gateway Endpoint

**What:** Route DynamoDB traffic through VPC gateway endpoint, not internet. Requires attaching to all private route tables in the spoke VPC.

```hcl
# Source: docs.aws.amazon.com/vpc/latest/privatelink/vpc-endpoints-ddb.html
resource "aws_vpc_endpoint" "dynamodb" {
  vpc_id            = aws_vpc.spoke.id
  service_name      = "com.amazonaws.${var.aws_region}.dynamodb"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = aws_route_table.private[*].id

  tags = merge(var.tags, { Name = "${var.project_name}-dynamodb-ep-${var.environment}" })
}
```
[VERIFIED: docs.aws.amazon.com/vpc, registry.terraform.io/providers/hashicorp/aws]

### Pattern 7: CosmosDB Private Endpoint Chain

**What:** Chain of four resources for fully private CosmosDB: account → private endpoint → DNS zone → DNS zone vnet link.

```hcl
# Source: learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-private-endpoints
resource "azurerm_cosmosdb_account" "main" {
  name                          = "${var.project_name}-cosmos-${var.environment}"
  location                      = var.location
  resource_group_name           = azurerm_resource_group.main.name
  offer_type                    = "Standard"
  kind                          = "GlobalDocumentDB"   # SQL API
  public_network_access_enabled = false

  consistency_policy {
    consistency_level = "Session"
  }

  geo_location {
    location          = var.location
    failover_priority = 0
  }

  tags = var.tags
}

resource "azurerm_private_endpoint" "cosmosdb" {
  name                = "${var.project_name}-cosmos-pe-${var.environment}"
  location            = var.location
  resource_group_name = azurerm_resource_group.main.name
  subnet_id           = var.private_subnet_id

  private_service_connection {
    name                           = "cosmosdb-psc"
    private_connection_resource_id = azurerm_cosmosdb_account.main.id
    subresource_names              = ["Sql"]
    is_manual_connection           = false
  }

  private_dns_zone_group {
    name                 = "cosmosdb-dns-group"
    private_dns_zone_ids = [azurerm_private_dns_zone.cosmosdb.id]
  }
}

resource "azurerm_private_dns_zone" "cosmosdb" {
  name                = "privatelink.documents.azure.com"
  resource_group_name = azurerm_resource_group.main.name
}

resource "azurerm_private_dns_zone_virtual_network_link" "cosmosdb" {
  name                  = "${var.project_name}-cosmos-dns-link-${var.environment}"
  resource_group_name   = azurerm_resource_group.main.name
  private_dns_zone_name = azurerm_private_dns_zone.cosmosdb.name
  virtual_network_id    = var.spoke_vnet_id
}
```
[CITED: learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-private-endpoints]

### Pattern 8: cert-manager Helm Release

**What:** Deploy cert-manager via OCI Helm chart with CRD installation.

```hcl
# Source: cert-manager.io/docs/installation/helm
resource "helm_release" "cert_manager" {
  name             = "cert-manager"
  repository       = "oci://quay.io/jetstack/charts"
  chart            = "cert-manager"
  version          = "v1.20.2"
  namespace        = "cert-manager"
  create_namespace = true
  wait             = true
  timeout          = 600

  set {
    name  = "crds.enabled"   # NOT installCRDs — deprecated since v1.15
    value = "true"
  }

  depends_on = [module.eks]   # or module.aks for Azure
}
```
[VERIFIED: cert-manager.io/docs/installation/helm — v1.20.2 OCI chart confirmed]

### Pattern 9: Secrets Store CSI Driver + Cloud Provider Helm Releases

**What:** Install the base CSI driver chart first, then the cloud-specific provider chart.

```hcl
# AWS
resource "helm_release" "secrets_store_csi" {
  name             = "secrets-store-csi-driver"
  repository       = "https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts"
  chart            = "secrets-store-csi-driver"
  version          = "1.6.0"
  namespace        = "kube-system"
  wait             = true
  timeout          = 300

  set { name = "syncSecret.enabled"; value = "true" }

  depends_on = [helm_release.cert_manager]
}

resource "helm_release" "csi_provider_aws" {
  name       = "csi-secrets-store-provider-aws"
  repository = "https://aws.github.io/secrets-store-csi-driver-provider-aws"
  chart      = "secrets-store-csi-driver-provider-aws"
  namespace  = "kube-system"
  wait       = true

  depends_on = [helm_release.secrets_store_csi]
}
```

```hcl
# Azure
resource "helm_release" "csi_provider_azure" {
  name             = "csi-secrets-store-provider-azure"
  repository       = "https://azure.github.io/secrets-store-csi-driver-provider-azure/charts"
  chart            = "csi-secrets-store-provider-azure"
  version          = "1.8.1"
  namespace        = "kube-system"
  wait             = true

  set { name = "secrets-store-csi-driver.install"; value = "false" }  # CSI driver already installed

  depends_on = [helm_release.secrets_store_csi]
}
```
[VERIFIED: kubernetes-sigs/secrets-store-csi-driver/releases — v1.6.0; artifacthub.io azure provider — 1.8.1]

### Anti-Patterns to Avoid

- **Static `aws_eks_cluster_auth` token in Helm/Kubernetes provider:** Token expires in 15 minutes; long `terraform apply` runs fail with 401. Use the `exec` block instead.
- **DynamoDB-based S3 lock:** Deprecated in Terraform 1.10+. Use `use_lockfile = true` on the S3 bucket instead.
- **`installCRDs = true` in cert-manager helm_release:** Deprecated since cert-manager v1.15. Use `crds.enabled = true`.
- **Applying Helm releases in the same Terraform root as cluster creation without explicit `depends_on`:** The Helm provider initializes at plan time and will fail if the cluster doesn't exist yet. Always add `depends_on = [aws_eks_node_group.main]` (not just the cluster, since nodes must be running for add-ons to deploy).
- **Hardcoded CIDR /24 subnets for EKS:** EKS needs enough IP space for pods; use /22 or larger for node subnets when using VPC CNI.
- **Omitting `aws_eks_addon` for vpc-cni, coredns, kube-proxy:** Without managed add-ons, these run as unmanaged DaemonSets that must be manually upgraded. Use `aws_eks_addon` with `depends_on = [aws_eks_node_group.main]`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| OIDC token exchange for IRSA | Custom IAM assume-role scripts | `aws_iam_openid_connect_provider` + annotated ServiceAccount | STS handles token validation; custom scripts cannot replicate the JWKS fetch |
| TLS certificate lifecycle | Self-signed cert renewal cron jobs | cert-manager `Certificate` + `ClusterIssuer` | Cert rotation is notoriously complex; cert-manager handles ACME, Private CA, validity windows |
| Secret injection into pods | Environment variable secrets in Deployment specs | Secrets Store CSI Driver `SecretProviderClass` | Env vars persist in pod spec YAML/etcd; CSI mounts are tmpfs in-memory and auto-rotate |
| Kubernetes cluster auth token refresh | Cached kubeconfig in CI | exec-based `aws eks get-token` in provider | Cached tokens expire in 15 minutes |
| S3 state lock | DynamoDB lock table (old pattern) | `use_lockfile = true` (S3 native) | DynamoDB lock is deprecated; S3 native uses conditional writes, no extra resource |
| Multi-region CosmosDB consistency | Application-level conflict resolution | CosmosDB Session consistency level | Session consistency gives read-your-writes with single-digit ms latency; conflict CRDT is complex |
| DNS resolution for private CosmosDB | /etc/hosts entries or Route53 private zone hacks | `azurerm_private_dns_zone` + vnet link | Azure requires `privatelink.documents.azure.com` zone registered in the VNet to resolve private IPs |

**Key insight:** All five "hard problems" here (OIDC exchange, cert rotation, secret injection, state locking, private DNS) have cloud-native or CNCF-standard solutions that encode years of operational knowledge. Custom implementations miss edge cases that only surface in production.

---

## Common Pitfalls

### Pitfall 1: EKS OIDC Thumbprint Drift

**What goes wrong:** The `thumbprint_list` in `aws_iam_openid_connect_provider` becomes stale when AWS rotates the EKS OIDC certificate. Terraform detects drift and wants to replace the OIDC provider, breaking existing IRSA roles.

**Why it happens:** The thumbprint is the SHA1 of the top intermediate CA cert. AWS rotates these. Hardcoding a static thumbprint value creates a time bomb.

**How to avoid:** Always use `data "tls_certificate"` to dynamically fetch the current thumbprint. For AWS-managed OIDC providers (EKS), AWS actually uses its trusted CA library and ignores the thumbprint, but the Terraform resource still requires it — fetch it dynamically to avoid diff noise.

**Warning signs:** Terraform plan shows replacement of `aws_iam_openid_connect_provider` despite no configuration changes.

[CITED: github.com/terraform-aws-modules/terraform-aws-eks/issues/2768]

### Pitfall 2: AKS Federated Credential Subject Case Sensitivity

**What goes wrong:** Federated credential token exchange silently fails if the Kubernetes `namespace` or `ServiceAccount` name case doesn't exactly match the `subject` field in `azurerm_federated_identity_credential`.

**Why it happens:** Azure Entra ID performs a case-sensitive string comparison on the `sub` claim from the OIDC token. `system:serviceaccount:My-App:my-sa` != `system:serviceaccount:my-app:my-sa`.

**How to avoid:** Use lowercase Kubernetes namespaces. Verify the ServiceAccount name in the cluster matches the Terraform variable exactly. The format is always `system:serviceaccount:<namespace>:<serviceaccount_name>`.

**Warning signs:** Pods return 401 on Azure SDK calls despite correct annotations; no error in Terraform apply.

[CITED: 2bcloud.io/using-aks-with-workload-identities-in-terraform]

### Pitfall 3: Helm Provider Kubernetes Cluster Unreachable

**What goes wrong:** `terraform apply` fails with "Kubernetes cluster unreachable" or "i/o timeout" when the `helm_release` resource is in the same root as the cluster resources.

**Why it happens:** Terraform resolves all provider configurations at plan time. The Helm/Kubernetes provider tries to connect to the cluster endpoint before the cluster exists.

**How to avoid:**
1. Add `depends_on = [aws_eks_node_group.main]` on every `helm_release` resource.
2. Alternatively (HashiCorp recommendation): Split cluster provisioning and in-cluster resource provisioning into separate Terraform roots with separate state files.

**Warning signs:** Error appears immediately at apply start, not after cluster creation.

[CITED: github.com/hashicorp/terraform-provider-helm/issues/400]

### Pitfall 4: aws_eks_addon Race with Node Group

**What goes wrong:** EKS add-ons (especially coredns) get stuck in DEGRADED state immediately after creation if the node group isn't fully ready.

**Why it happens:** CoreDNS pods can't schedule until nodes are Ready. The add-on API marks them degraded, and the Terraform resource can time out.

**How to avoid:** Always add `depends_on = [aws_eks_node_group.main]` to every `aws_eks_addon` resource.

**Warning signs:** `terraform apply` hangs for 20+ minutes on `aws_eks_addon.coredns`; AWS Console shows DEGRADED status.

[CITED: github.com/terraform-aws-modules/terraform-aws-eks/issues/1801]

### Pitfall 5: AKS Federated Credential Propagation Delay

**What goes wrong:** Immediately after `terraform apply`, pods fail to authenticate to Azure services even though the federated credential was created successfully.

**Why it happens:** Entra ID takes 10–60 seconds to propagate a newly created federated identity credential across all regions.

**How to avoid:** This is unavoidable but expected. Document it so CI pipelines don't immediately run integration tests; add a brief wait or retry in test harnesses.

**Warning signs:** First pod startup after fresh deploy fails; second attempt succeeds.

[CITED: 2bcloud.io/using-aks-with-workload-identities-in-terraform]

### Pitfall 6: CosmosDB + public_network_access_enabled = false Bootstrapping

**What goes wrong:** Setting `public_network_access_enabled = false` on `azurerm_cosmosdb_account` and then trying to create the private endpoint in the same `terraform apply` can succeed, but a subsequent `terraform plan` may show spurious diffs or the Terraform `azurerm` provider may error when refreshing state if the machine running Terraform has no private network access to CosmosDB.

**Why it happens:** The `azurerm` provider reads resource state via the Azure management plane API (not the data plane), so this is less of an issue than it seems — but the Terraform state refresh for CosmosDB properties still goes through ARM, not the private endpoint. The main risk is that Portal-based connectivity checks (for debugging) will fail after setting this flag.

**How to avoid:** Set `public_network_access_enabled = false` from the start; do not toggle it after the fact. Run Terraform from a machine or CI runner that has Azure management plane (ARM) access (which is always public).

[CITED: github.com/hashicorp/terraform-provider-azurerm/issues/18450]

### Pitfall 7: S3 State Bucket Terraform CLI Version Mismatch

**What goes wrong:** The local Terraform CLI is v1.9.8. `use_lockfile = true` was introduced in Terraform 1.10. Running `terraform init` with v1.9.8 and a backend config containing `use_lockfile = true` will fail with an unknown argument error.

**Why it happens:** The local dev environment has v1.9.8, but the feature is a v1.10+ addition.

**How to avoid:** Upgrade the local Terraform CLI to >= 1.10 before writing bootstrap code. Use `required_version = ">= 1.10"` in all modules to enforce this at init time.

**Warning signs:** `Error: Unsupported argument` on `terraform init` for the S3 backend block.

[VERIFIED: terraform v1.9.8 confirmed installed locally; use_lockfile verified as v1.10+ feature]

---

## Code Examples

### DynamoDB Table (On-Demand + Encryption)

```hcl
# Source: registry.terraform.io/providers/hashicorp/aws docs
resource "aws_dynamodb_table" "main" {
  name         = "${var.project_name}-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"
  range_key    = "sk"

  attribute {
    name = "pk"
    type = "S"
  }
  attribute {
    name = "sk"
    type = "S"
  }

  server_side_encryption {
    enabled = true
    # kms_key_arn = optional — defaults to AWS-managed key
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = var.tags
}

output "database_endpoint" {
  value = aws_dynamodb_table.main.name
}
output "resource_id" {
  value = aws_dynamodb_table.main.arn
}
```

### AKS Cluster with Workload Identity (full)

```hcl
# Source: learn.microsoft.com/en-us/azure/aks/workload-identity-deploy-cluster
resource "azurerm_kubernetes_cluster" "main" {
  name                      = "${var.project_name}-aks-${var.environment}"
  location                  = var.location
  resource_group_name       = var.resource_group_name
  dns_prefix                = "${var.project_name}-${var.environment}"
  kubernetes_version        = "1.33"   # pin to latest stable supported by AKS
  oidc_issuer_enabled       = true
  workload_identity_enabled = true

  default_node_pool {
    name           = "system"
    node_count     = var.node_count
    vm_size        = "Standard_D2s_v3"
    vnet_subnet_id = var.subnet_id
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin = "azure"
    network_policy = "azure"
  }

  tags = var.tags
}
```

### EKS Managed Node Group Pattern

```hcl
# Source: docs.aws.amazon.com/eks/latest/userguide/managed-node-groups.html
resource "aws_eks_cluster" "main" {
  name     = "${var.project_name}-eks-${var.environment}"
  version  = "1.33"   # pin to latest EKS-supported version
  role_arn = aws_iam_role.eks_cluster.arn

  vpc_config {
    subnet_ids              = var.private_subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = false   # private cluster
  }

  tags = var.tags
  depends_on = [aws_iam_role_policy_attachment.eks_cluster_policy]
}

resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "${var.project_name}-ng-${var.environment}"
  node_role_arn   = aws_iam_role.eks_node.arn
  subnet_ids      = var.private_subnet_ids

  scaling_config {
    desired_size = var.node_count   # 2
    min_size     = var.node_count
    max_size     = var.node_count + 2
  }

  instance_types = ["t3.medium"]

  tags = var.tags
  depends_on = [
    aws_iam_role_policy_attachment.eks_node_policy,
    aws_iam_role_policy_attachment.eks_cni_policy,
    aws_iam_role_policy_attachment.eks_ecr_policy,
  ]
}

resource "aws_eks_addon" "vpc_cni" {
  cluster_name      = aws_eks_cluster.main.name
  addon_name        = "vpc-cni"
  resolve_conflicts_on_create = "OVERWRITE"
  depends_on        = [aws_eks_node_group.main]
}

resource "aws_eks_addon" "coredns" {
  cluster_name      = aws_eks_cluster.main.name
  addon_name        = "coredns"
  resolve_conflicts_on_create = "OVERWRITE"
  depends_on        = [aws_eks_node_group.main]
}

resource "aws_eks_addon" "kube_proxy" {
  cluster_name      = aws_eks_cluster.main.name
  addon_name        = "kube-proxy"
  resolve_conflicts_on_create = "OVERWRITE"
  depends_on        = [aws_eks_node_group.main]
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| S3 backend + DynamoDB lock table | S3 backend + `use_lockfile = true` | Terraform 1.10 (Nov 2024) | Removes the need for a separate DynamoDB table in bootstrap; DynamoDB lock deprecated |
| `installCRDs = true` in cert-manager | `crds.enabled = true` | cert-manager v1.15 (2024) | Old flag still works but emits deprecation warning; v1.20.2 OCI chart is standard |
| `aws_eks_cluster_auth` static token | Helm/K8s provider `exec` block | Ongoing (became critical with longer apply times) | Prevents 15-minute token expiry during large applies |
| Kubernetes provider 2.x | Kubernetes provider 3.x | 2024 | Requires Terraform >= 1.0; plugin protocol v6 |
| Helm provider 2.x | Helm provider 3.x | 2024 | Requires Terraform >= 1.0 |
| AWS provider 5.x | AWS provider 6.x | June 2025 | Per-resource `region` attribute; multi-region support without provider aliases |

**Deprecated/outdated:**
- `aws_dynamodb_table` as state lock resource: Was the standard before Terraform 1.10. Now deprecated. The bootstrap module should still create one for workload data but NOT for state locking.
- `installCRDs` in cert-manager helm values: Use `crds.enabled` instead.
- `data "aws_eks_cluster_auth"` as static token for providers: Use exec block.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | EKS supports Kubernetes 1.33 as a standard-support version in May 2026 | Code Examples | If EKS has moved to 1.34/1.35 as minimum, the pinned version may not be available — must verify at plan time with `aws eks describe-addon-versions` |
| A2 | AKS supports Kubernetes 1.33 in May 2026 | Code Examples | AKS version support changes monthly; must verify with `az aks get-versions --location <region>` at plan time |
| A3 | AWS provider v6 requires Terraform >= 1.0 (not 1.10) | Standard Stack | If minimum is higher, the required_version constraint needs updating |
| A4 | `csi-secrets-store-provider-azure` chart v1.8.1 is the latest stable | Standard Stack | Chart versions for CNCF projects move fast; should verify at `helm search repo` time |
| A5 | Hub-Spoke with VPC Peering (not Transit Gateway) is sufficient for this PoC | Architecture | Landing zone spec says Hub-Spoke but doesn't specify TGW vs peering; VPC peering is simpler and cheaper for 2-VPC PoC — if TGW is explicitly required, cost and complexity increase |

---

## Open Questions

1. **Kubernetes version to pin for EKS and AKS**
   - What we know: EKS supports up to 1.35 as of Jan 2026; AKS supports 1.33–1.35 in May 2026
   - What's unclear: What is the exact latest-stable version available in the target regions at plan time?
   - Recommendation: Omit the `kubernetes_version` pin from the module default and set it in `.tfvars` per environment; add a comment to check `aws eks describe-addon-versions` and `az aks get-versions` before first apply.

2. **Hub-Spoke implementation: VPC Peering vs Transit Gateway**
   - What we know: Landing-zones spec says Hub-Spoke with Hub VPC centralizing NAT/inspection; spec does not explicitly state Transit Gateway
   - What's unclear: For a 2-VPC PoC (1 hub + 1 spoke per environment), VPC Peering is simpler, cheaper, and sufficient. Transit Gateway adds ~$0.05/hour/attachment + data transfer costs.
   - Recommendation: Use VPC Peering for the PoC. Transit Gateway is warranted only when there are 3+ VPCs needing transitive routing. If the intent is to demonstrate enterprise patterns, flag for user confirmation.

3. **Target AWS region and Azure region**
   - What we know: Not specified in CONTEXT.md
   - What's unclear: Required for service name in Gateway Endpoint (`com.amazonaws.<region>.dynamodb`) and AKS node pool availability zones
   - Recommendation: Add `aws_region` and `azure_location` as required variables in all live roots; no defaults.

4. **Separate resource groups per module or single RG for Azure**
   - What we know: CONTEXT.md doesn't specify
   - What's unclear: AKS, CosmosDB, and VNet could share one RG per environment, or each could have its own
   - Recommendation: Use one resource group per environment per cloud (e.g., `cell-arch-dev`) for simplicity; IAM/RBAC boundaries can use Azure RBAC at the subscription level.

5. **CosmosDB database and container names**
   - What we know: `cmd/sidecar/main.go` reads `COSMOS_DATABASE` and `COSMOS_CONTAINER` env vars
   - What's unclear: Exact names for dev/hom/prod environments
   - Recommendation: Set as variables in `.tfvars` with sensible defaults (`cell-arch-db`, `items`).

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Terraform CLI | All `terraform` commands | Partial | 1.9.8 (need >= 1.10) | Download 1.15.1 from releases.hashicorp.com |
| Azure CLI (`az`) | Azure bootstrap, AKS auth | Yes | 2.85.0 | — |
| AWS CLI (`aws`) | EKS exec auth in Helm/K8s provider | No | — | Install from aws.amazon.com/cli |
| `helm` CLI | Manual chart inspection/debug | No | — | Not required for Terraform (uses helm provider) |
| `kubectl` | Manual cluster debugging | No | — | Not required for Terraform execution |
| Go | Application code (other phases) | No | — | Not required for Phase 3 |

**Missing dependencies with no fallback:**
- **Terraform >= 1.10:** Current v1.9.8 cannot use `use_lockfile = true`. Must upgrade before bootstrap.
- **AWS CLI:** Required by the Helm/Kubernetes provider `exec` block (`aws eks get-token`). Must be installed on any machine that runs `terraform apply` for AWS live environments.

**Missing dependencies with fallback:**
- `helm` CLI: Not needed for Terraform execution; only for ad-hoc chart debugging.
- `kubectl`: Not needed for Terraform execution.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | `go test` (Go standard) + testcontainers-go for integration |
| Config file | none for Phase 3 (IaC-only; no Go test files in this phase) |
| Quick run command | `terraform validate && terraform plan -var-file=dev.tfvars` |
| Full suite command | `terraform validate && terraform plan -var-file=dev.tfvars -out=tfplan && terraform show -json tfplan \| jq ...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TERR-01 | VPC with private subnets only, Gateway Endpoint present | Terraform plan + tfsec | `terraform plan -var-file=dev.tfvars` (check no public subnet resources) | ❌ Wave 0 |
| TERR-02 | VNet with private subnets, Private Link DNS zone | Terraform plan + tfsec | `terraform plan -var-file=dev.tfvars` | ❌ Wave 0 |
| TERR-03 | EKS cluster OIDC provider + IAM trust policy | Terraform plan output | `terraform plan -var-file=dev.tfvars` | ❌ Wave 0 |
| TERR-04 | AKS oidc_issuer_enabled + workload_identity + federated cred | Terraform plan output | `terraform plan -var-file=dev.tfvars` | ❌ Wave 0 |
| TERR-05 | DynamoDB PAY_PER_REQUEST + SSE enabled | Terraform plan + tfsec | `tfsec infra/modules/aws-dynamodb` | ❌ Wave 0 |
| TERR-06 | CosmosDB Session consistency + no public access | Terraform plan output | `terraform plan -var-file=dev.tfvars` | ❌ Wave 0 |
| TERR-07 | S3 backend use_lockfile / azurerm Blob Lease | Bootstrap plan output | `terraform -chdir=bootstrap/aws plan` | ❌ Wave 0 |
| TERR-08 | Symmetric module structure, required outputs present | `terraform output` after apply | manual verification or output validation script | ❌ Wave 0 |
| SECR-01 | CSI driver + provider Helm releases succeed | helm_release plan | `terraform plan -var-file=dev.tfvars` (check release resources) | ❌ Wave 0 |
| SECR-02 | cert-manager Helm release with crds.enabled=true | helm_release plan | `terraform plan -var-file=dev.tfvars` | ❌ Wave 0 |
| SECR-03 | No public subnets in any module | tfsec + plan | `tfsec infra/modules/` | ❌ Wave 0 |

**Note:** Phase 3 is IaC-only (no Go code). Validation is `terraform validate`, `terraform plan`, and `tfsec`. All test infrastructure must be created in Wave 0 of the plan.

### Sampling Rate

- **Per task commit:** `terraform validate && terraform fmt -check`
- **Per wave merge:** `terraform validate && tfsec infra/modules/ && terraform plan -var-file=dev.tfvars`
- **Phase gate:** Full plan green (no errors) for all 3 environments × 2 clouds before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] `infra/bootstrap/aws/main.tf` — TERR-07 (S3 state backend resources)
- [ ] `infra/bootstrap/azure/main.tf` — TERR-07 (Blob state backend resources)
- [ ] `infra/modules/aws-vpc/` — TERR-01 (VPC + Gateway Endpoint)
- [ ] `infra/modules/aws-eks/` — TERR-03 (EKS + IRSA)
- [ ] `infra/modules/aws-dynamodb/` — TERR-05
- [ ] `infra/modules/azure-vnet/` — TERR-02
- [ ] `infra/modules/azure-aks/` — TERR-04
- [ ] `infra/modules/azure-cosmosdb/` — TERR-06
- [ ] `infra/live/aws/dev/` + `hom/` + `prod/` — TERR-08, SECR-01, SECR-02
- [ ] `infra/live/azure/dev/` + `hom/` + `prod/` — TERR-08, SECR-01, SECR-02
- [ ] `.tfsec.yaml` or `tfsec` config — SECR-03 validation
- [ ] `infra/live/aws/dev/dev.tfvars` sample — seeding environment variables

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | Yes | IRSA (AWS) + Workload Identity (Azure) — no static credentials |
| V3 Session Management | No | IaC phase; no session tokens in Terraform state |
| V4 Access Control | Yes | IAM policies scoped to specific DynamoDB table; Cosmos RBAC on container |
| V5 Input Validation | No | Terraform variables are dev-facing config, not user input |
| V6 Cryptography | Yes | DynamoDB SSE with AWS-managed key; CosmosDB encrypted at rest by default; TLS 1.2+ enforced by Azure on CosmosDB private link |

### Known Threat Patterns for this Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Static AWS/Azure credentials in Terraform state | Information Disclosure | IRSA + Workload Identity; no `access_key`/`secret_key` in provider blocks |
| Public DynamoDB endpoint (outside VPC) | Spoofing / Elevation | Gateway VPC Endpoint; IAM policy requiring `aws:SourceVpc` condition [ASSUMED] |
| CosmosDB public internet access | Information Disclosure | `public_network_access_enabled = false` + private endpoint |
| Helm chart from unauthenticated repo | Tampering | Pin chart versions; use OCI for cert-manager (signed images) |
| Overly broad IRSA trust policy | Elevation of Privilege | `StringEquals` condition on both `:sub` and `:aud`; no wildcards |
| AKS node pool running with admin privileges | Elevation of Privilege | `only_critical_addons_enabled = false` for system pool; user workloads in system pool is acceptable for PoC scale |
| Terraform state file contains sensitive outputs | Information Disclosure | S3 server-side encryption enabled; Azure Blob encryption at rest by default |

---

## Sources

### Primary (HIGH confidence)
- `developer.hashicorp.com/terraform/language/backend/s3` — S3 native locking, use_lockfile behavior, DynamoDB deprecation
- `cert-manager.io/docs/installation/helm` — v1.20.2, OCI chart URL, crds.enabled parameter
- `learn.microsoft.com/en-us/azure/aks/workload-identity-deploy-cluster` — AKS Workload Identity Terraform pattern
- `docs.aws.amazon.com/vpc/latest/privatelink/vpc-endpoints-ddb.html` — DynamoDB Gateway Endpoint
- `learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-private-endpoints` — CosmosDB Private Link chain

### Secondary (MEDIUM confidence)
- `releases.hashicorp.com/terraform-provider-aws` — v6.44 latest confirmed (WebSearch verified)
- `releases.hashicorp.com/terraform-provider-azurerm` — v4.71 latest confirmed (WebSearch verified)
- `releases.hashicorp.com/terraform-provider-kubernetes` — v3.1.0 latest (WebSearch verified)
- `releases.hashicorp.com/terraform-provider-helm` — v3.1.1 latest (WebSearch verified)
- `releases.hashicorp.com/terraform` — v1.15.1 CLI latest (WebSearch verified)
- `github.com/kubernetes-sigs/secrets-store-csi-driver/releases` — v1.6.0 released Apr 2026 (WebFetch verified)
- `artifacthub.io/packages/helm/csi-secrets-store-provider-azure` — v1.8.1 (WebFetch verified)
- `github.com/terraform-aws-modules/terraform-aws-eks/issues/2768` — thumbprint drift issue
- `github.com/terraform-aws-modules/terraform-aws-eks/issues/1801` — add-on race with node group
- `github.com/hashicorp/terraform-provider-helm/issues/400` — cluster unreachable race condition

### Tertiary (LOW confidence)
- `2bcloud.io/using-aks-with-workload-identities-in-terraform` — federated credential pitfalls (community article, not official docs)

---

## Metadata

**Confidence breakdown:**
- Standard stack (provider versions): HIGH — verified from official release pages and WebSearch
- Architecture patterns (EKS IRSA, AKS Workload Identity): HIGH — verified against official AWS and Microsoft docs
- Helm chart versions (cert-manager, CSI driver): HIGH — verified from official project releases
- State backend (S3 native locking): HIGH — verified from official Terraform docs
- Common pitfalls: MEDIUM — mix of official GitHub issues (HIGH) and community articles (MEDIUM)
- CIDR recommendations: LOW — not researched; delegated to Claude's discretion

**Research date:** 2026-05-06
**Valid until:** 2026-08-06 (90 days) — provider versions and Kubernetes versions change frequently; re-verify Kubernetes version pins before executing

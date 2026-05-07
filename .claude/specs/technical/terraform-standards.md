# Infrastructure Spec: Terraform IaC Standardization

## 1. State Management and Backends

* **Backend Symmetry**: Use `s3` backend with `use_lockfile = true` (S3 native locking) for AWS and `azurerm` backend with automatic Blob Lease locking for Azure.
* **DynamoDB Lock Deprecated**: DynamoDB-based S3 locking is deprecated in Terraform 1.10+. Use `use_lockfile = true` on the S3 backend instead. No DynamoDB lock table is needed for state locking in bootstrap.
* **Per-Environment Isolation**: Each environment (dev, hom, prod) has its own isolated backend. AWS: own S3 bucket + use_lockfile. Azure: own Blob container. Full blast-radius isolation.
* **Bootstrap Pattern**: A local-state bootstrap module at `infra/bootstrap/{aws,azure}/` creates the remote backends before the live roots are initialized. Bootstrap runs once per environment (dev, hom, prod).
* **Minimum Terraform CLI**: `>= 1.10` required for `use_lockfile = true`. Pin `required_version = ">= 1.10"` in all modules.

## 2. Module Structure

* **Encapsulation**: Cloud resources (EKS, AKS, DynamoDB, CosmosDB, VPC, VNet) are encapsulated in reusable modules under `infra/modules/`.
* **Live Environment Roots**: Live environments under `infra/live/{aws,azure}/{dev,hom,prod}/`. Each live root has its own `providers.tf` with pinned provider versions.
* **Standardized Outputs**: Each database module must expose: `database_endpoint`, `resource_id`, `iam_policy_arn` (AWS) or `policy_object_id` (Azure). The `azure-cosmosdb` module additionally exposes `account_name` for use in `azurerm_cosmosdb_sql_role_assignment`.
* **Mandatory Tags**: All resources carry: `project_name`, `managed_by: terraform`, `environment: {dev|hom|prod}`.

## 3. Environment Strategy

* **Three Environments**: `dev`, `hom` (homologation/staging), `prod`.
* **Per-Environment tfvars**: Environment-specific values in `.tfvars` files per environment per cloud (e.g., `dev.tfvars`, `hom.tfvars`, `prod.tfvars`).
* **Naming Convention**: `{project_name}-tfstate-{cloud}-{env}` for state backend resources (e.g., `cell-arch-tfstate-aws-dev`, `cell-arch-tfstate-azure-dev`).

## 4. Provider Version Pins

| Provider | Version | Purpose |
|----------|---------|---------|
| hashicorp/aws | `~> 6.0` | All AWS resources |
| hashicorp/azurerm | `~> 4.0` | All Azure resources |
| hashicorp/kubernetes | `~> 3.1` | Post-cluster Kubernetes resources |
| hashicorp/helm | `~> 3.1` | Helm releases (cert-manager, CSI) |
| hashicorp/tls | `~> 4.0` | EKS OIDC thumbprint data source |

## 5. Kubernetes/Helm Provider Authentication

Use exec-based authentication to avoid the 15-minute static token expiry:

```hcl
# AWS -- aws eks get-token
exec {
  api_version = "client.authentication.k8s.io/v1beta1"
  command     = "aws"
  args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name, "--region", var.aws_region]
}

# Azure -- kubelogin with Azure CLI login
exec {
  api_version = "client.authentication.k8s.io/v1beta1"
  command     = "kubelogin"
  args        = ["get-token", "--login", "azurecli", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"]
}
```

Never use `data "aws_eks_cluster_auth"` or static kubeconfig tokens. These expire in 15 minutes and break long `terraform apply` runs.
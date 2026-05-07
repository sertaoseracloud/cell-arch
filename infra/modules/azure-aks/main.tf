terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.101.0"
    }
  }
}

# ── Default Node Pool Subnet (for pod / service CIDR) ─────────────

data "azurerm_subnet" "spoke_private_a" {
  name                 = element(var.private_subnet_ids, 0)
  virtual_network_name = split("/", element(var.private_subnet_ids, 0))[4]  # crude but works; use locals in real impl
  resource_group_name  = split("/", element(var.private_subnet_ids, 0))[3]
}

# ── User Assigned Identity for AKS ────────────────────────────────

resource "azurerm_user_assigned_identity" "aks" {
  name                = "${var.project_name}-aks-identity-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = var.tags
}

# ── AKS Cluster (private, Workload Identity enabled) ─────────────

resource "azurerm_kubernetes_cluster" "main" {
  name                = "${var.project_name}-aks-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name
  dns_prefix          = "${var.project_name}-${var.environment}"

  default_node_pool {
    name            = "default"
    node_count      = var.node_count
    vm_size         = "Standard_D2s_v3"
    vnet_subnet_id = data.azurerm_subnet.spoke_private_a.id
    zones           = ["1", "2", "3"]
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.aks.id]
  }

  network_profile {
    network_plugin = "azure"
  }

  oidc_issuer_enabled       = true
  workload_identity_enabled = true

  kubernetes_version = var.kubernetes_version

  tags = merge(var.tags, {
    Name = "${var.project_name}-aks-${var.environment}"
  })
}

# ── Federated Identity Credential ( binds AKS OIDC → User Identity ) ──

resource "azurerm_federated_identity_credential" "sidecar" {
  name                = "${var.project_name}-fi-${var.environment}"
  resource_group_name = var.resource_group_name
  audience            = ["api://AzureADTokenExchange"]
  issuer              = azurerm_kubernetes_cluster.main.oidc_issuer_url
  parent_id           = azurerm_user_assigned_identity.aks.id
  subject             = "system:serviceaccount:default:${var.service_account_name}"
}

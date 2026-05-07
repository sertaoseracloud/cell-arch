terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.101.0"
    }
  }
}

# ── Hub VNet ────────────────────────────────────────────
# CIDR 10.2.0.0/16 (65k IPs) — Azure Hub for shared services

resource "azurerm_virtual_network" "hub" {
  name                = "${var.project_name}-hub-vnet-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name
  address_space       = ["10.2.0.0/16"]
  tags                = merge(var.tags, {
    Name = "${var.project_name}-hub-vnet-${var.environment}"
  })
}

# Hub private subnet (for firewall, etc.)
resource "azurerm_subnet" "hub_private" {
  name                 = "${var.project_name}-hub-private-${var.environment}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.hub.name
  address_prefixes     = ["10.2.0.0/22"]
}

# ── Spoke VNet ────────────────────────────────────────────
# CIDR 10.3.0.0/16 (65k IPs) — AKS nodes in private subnets

resource "azurerm_virtual_network" "spoke" {
  name                = "${var.project_name}-spoke-vnet-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name
  address_space       = ["10.3.0.0/16"]
  tags                = merge(var.tags, {
    Name = "${var.project_name}-spoke-vnet-${var.environment}"
  })
}

# Spoke private subnets (AKS nodes — no public IPs)
resource "azurerm_subnet" "spoke_private_a" {
  name                 = "${var.project_name}-spoke-private-a-${var.environment}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.spoke.name
  address_prefixes     = ["10.3.0.0/22"]
}

resource "azurerm_subnet" "spoke_private_b" {
  name                 = "${var.project_name}-spoke-private-b-${var.environment}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.spoke.name
  address_prefixes     = ["10.3.4.0/22"]
}

# ── VNet Peering (Hub ↔ Spoke) ────────────────────────────────

resource "azurerm_virtual_network_peering" "hub_to_spoke" {
  name                      = "${var.project_name}-hub-to-spoke-${var.environment}"
  resource_group_name       = var.resource_group_name
  virtual_network_name      = azurerm_virtual_network.hub.name
  remote_virtual_network_id = azurerm_virtual_network.spoke.id
  allow_virtual_network_access = true
}

resource "azurerm_virtual_network_peering" "spoke_to_hub" {
  name                      = "${var.project_name}-spoke-to-hub-${var.environment}"
  resource_group_name       = var.resource_group_name
  virtual_network_name      = azurerm_virtual_network.spoke.name
  remote_virtual_network_id = azurerm_virtual_network.hub.id
  allow_virtual_network_access = true
}

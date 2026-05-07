terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.101.0"
    }
  }
}

# ── Hub VNet ───────────────────────────────────────────
# CIDR 10.10.0.0/16 — Azure Hub for shared services

resource "azurerm_virtual_network" "hub" {
  name                = "${var.project_name}-hub-vnet-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name
  address_space       = ["10.10.0.0/16"]
  tags                = merge(var.tags, {
    Name = "${var.project_name}-hub-vnet-${var.environment}"
  })
}

resource "azurerm_subnet" "hub_firewall" {
  name                 = "AzureFirewallSubnet"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.hub.name
  address_prefixes     = ["10.10.0.0/26"]
}

# ── Spoke VNet ────────────────────────────────────────────
# CIDR 10.11.0.0/16 — AKS nodes in private subnets

resource "azurerm_virtual_network" "spoke" {
  name                = "${var.project_name}-spoke-vnet-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name
  address_space       = ["10.11.0.0/16"]
  tags                = merge(var.tags, {
    Name = "${var.project_name}-spoke-vnet-${var.environment}"
  })
}

resource "azurerm_subnet" "aks" {
  name                 = "${var.project_name}-aks-subnet-${var.environment}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.spoke.name
  address_prefixes     = ["10.11.0.0/22"]
}

resource "azurerm_subnet" "privatelink" {
  name                 = "${var.project_name}-pe-subnet-${var.environment}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.spoke.name
  address_prefixes     = ["10.11.8.0/27"]
}

# ── VNet Peering (Hub ↔ Spoke) ────────────────────────────────

resource "azurerm_virtual_network_peering" "hub_to_spoke" {
  name                      = "${var.project_name}-hub-to-spoke-${var.environment}"
  resource_group_name       = var.resource_group_name
  virtual_network_name      = azurerm_virtual_network.hub.name
  remote_virtual_network_id = azurerm_virtual_network.spoke.id
  allow_virtual_network_access = true
  allow_forwarded_traffic      = true
  allow_gateway_transit        = false
}

resource "azurerm_virtual_network_peering" "spoke_to_hub" {
  name                      = "${var.project_name}-spoke-to-hub-${var.environment}"
  resource_group_name       = var.resource_group_name
  virtual_network_name      = azurerm_virtual_network.spoke.name
  remote_virtual_network_id = azurerm_virtual_network.hub.id
  allow_virtual_network_access = true
  allow_forwarded_traffic      = true
  use_remote_gateways          = false
}

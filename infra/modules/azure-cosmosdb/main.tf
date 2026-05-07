terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.101.0"
    }
  }
}

# ── CosmosDB Account ────────────────────────────────────────

resource "azurerm_cosmosdb_account" "main" {
  name                = "${var.project_name}-cosmos-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name
  offer_type          = "Standard"
  kind                = "GlobalDocumentDB"

  consistency_policy {
    consistency_level = "Session"
  }

  geo_location {
    location        = var.location
    failover_priority = 0
  }

  public_network_access_enabled = false

  tags = merge(var.tags, {
    Name = "${var.project_name}-cosmos-${var.environment}"
  })
}

# ── SQL Database ────────────────────────────────────────────

resource "azurerm_cosmosdb_sql_database" "main" {
  name                = "${var.project_name}-db-${var.environment}"
  resource_group_name = var.resource_group_name
  account_name        = azurerm_cosmosdb_account.main.name
}

# ── SQL Container ───────────────────────────────────────────

resource "azurerm_cosmosdb_sql_container" "main" {
  name                = "${var.project_name}-container-${var.environment}"
  resource_group_name = var.resource_group_name
  account_name        = azurerm_cosmosdb_account.main.name
  database_name       = azurerm_cosmosdb_sql_database.main.name
  partition_key_path = "/task"
}

# ── Private Endpoint ─────────────────────────────────────────

resource "azurerm_private_endpoint" "cosmosdb" {
  name                = "${var.project_name}-cosmos-ep-${var.environment}"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.endpoint_subnet_id

  private_service_connection {
    name                           = "${var.project_name}-cosmos-psc-${var.environment}"
    private_connection_resource_id = azurerm_cosmosdb_account.main.id
    is_manual_connection           = false
  }

  tags = var.tags
}

# ── DNS Zone Link ───────────────────────────────────────────

resource "azurerm_private_dns_zone_virtual_network_link" "cosmosdb" {
  name                  = "${var.project_name}-cosmos-dns-link-${var.environment}"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name  = var.private_dns_zone_name
  virtual_network_id    = var.vnet_id
}

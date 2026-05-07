module "vnet" {
  source = "../../../modules/azure-vnet"

  project_name = var.project_name
  environment = var.environment
  location     = var.location
  tags        = var.tags
}

module "aks" {
  source = "../../../modules/azure-aks"

  project_name       = var.project_name
  environment       = var.environment
  location           = var.location
  private_subnet_ids = module.vnet.spoke_private_subnet_ids
  kubernetes_version = "1.29"
  node_count         = 2
  tags              = var.tags
}

module "cosmosdb" {
  source = "../../../modules/azure-cosmosdb"

  project_name     = var.project_name
  environment     = var.environment
  location         = var.location
  endpoint_subnet_id = element(module.vnet.spoke_private_subnet_ids, 0)
  private_dns_zone_id = azurerm_private_dns_zone.cosmos[0].id
  tags             = var.tags
}

# Private DNS zone for CosmosDB (hub vnet link)
resource "azurerm_private_dns_zone" "cosmos" {
  count               = 1
  name                = "privatelink.documents.azure.com"
  resource_group_name = var.resource_group_name
  tags                = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "hub" {
  count                 = 1
  name                  = "${var.project_name}-cosmos-dns-link-${var.environment}"
  resource_group_name   = var.resource_group_name
  private_dns_zone_id  = azurerm_private_dns_zone.cosmos[0].id
  virtual_network_id    = module.vnet.hub_vnet_id
  tags                  = var.tags
}

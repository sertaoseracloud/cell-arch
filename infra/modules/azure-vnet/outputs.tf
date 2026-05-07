output "hub_vnet_id" {
  description = "ID of the Hub VNet"
  value       = azurerm_virtual_network.hub.id
}

output "spoke_vnet_id" {
  description = "ID of the Spoke VNet (AKS lives here)"
  value       = azurerm_virtual_network.spoke.id
}

output "aks_subnet_id" {
  description = "AKS subnet ID"
  value       = azurerm_subnet.aks.id
}

output "privatelink_subnet_id" {
  description = "Private endpoint subnet ID"
  value       = azurerm_subnet.privatelink.id
}

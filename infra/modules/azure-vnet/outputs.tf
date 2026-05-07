output "hub_vnet_id" {
  description = "ID of the Hub VNet"
  value       = azurerm_virtual_network.hub.id
}

output "spoke_vnet_id" {
  description = "ID of the Spoke VNet (AKS lives here)"
  value       = azurerm_virtual_network.spoke.id
}

output "spoke_private_subnet_ids" {
  description = "Private subnet IDs in the Spoke VNet — passed to azure-aks module"
  value       = [azurerm_subnet.spoke_private_a.id, azurerm_subnet.spoke_private_b.id]
}

output "cosmosdb_endpoint" {
  description = "CosmosDB account endpoint URL — used as AZURE_COSMOS_ENDPOINT env var"
  value       = azurerm_cosmosdb_account.main.endpoint
}

output "database_name" {
  description = "SQL database name — used as COSMOS_DATABASE env var"
  value       = azurerm_cosmosdb_sql_database.main.name
}

output "container_name" {
  description = "SQL container name — used as COSMOS_CONTAINER env var"
  value       = azurerm_cosmosdb_sql_container.main.name
}

output "resource_id" {
  description = "CosmosDB account resource ID"
  value       = azurerm_cosmosdb_account.main.id
}

output "policy_object_id" {
  description = "Placeholder — Azure uses RBAC Role Assignments, not policy objects"
  value       = azurerm_cosmosdb_account.main.id
}

output "aks_cluster_name" {
  description = "AKS cluster name"
  value       = module.aks.cluster_name
}

output "cosmosdb_endpoint" {
  description = "CosmosDB account endpoint"
  value       = module.cosmosdb.cosmosdb_endpoint
}

output "cosmosdb_database" {
  description = "SQL database name"
  value       = module.cosmosdb.database_name
}

output "cosmosdb_container" {
  description = "SQL container name"
  value       = module.cosmosdb.container_name
}

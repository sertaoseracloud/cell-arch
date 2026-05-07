output "hub_vpc_id" {
  description = "ID of the Hub VPC"
  value       = aws_vpc.hub.id
}

output "spoke_vpc_id" {
  description = "ID of the Spoke VPC (EKS lives here)"
  value       = aws_vpc.spoke.id
}

output "spoke_private_subnet_ids" {
  description = "Private subnet IDs in the Spoke VPC — passed to aws-eks module"
  value       = [aws_subnet.spoke_private_a.id, aws_subnet.spoke_private_b.id]
}

output "spoke_route_table_ids" {
  description = "Route table IDs for Spoke private subnets (for additional endpoint attachments)"
  value       = aws_route_table.spoke_private[*].id
}

output "dynamodb_endpoint_id" {
  description = "ID of the DynamoDB Gateway VPC Endpoint"
  value       = aws_vpc_endpoint.dynamodb.id
}

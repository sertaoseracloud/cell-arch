output "database_endpoint" {
  description = "DynamoDB table name — used as DYNAMODB_TABLE env var in sidecar"
  value       = aws_dynamodb_table.main.name
}

output "resource_id" {
  description = "DynamoDB table ARN"
  value       = aws_dynamodb_table.main.arn
}

output "iam_policy_arn" {
  description = "ARN of the IAM policy granting DynamoDB CRUD — attach to IRSA role in live root"
  value       = aws_iam_policy.dynamodb_access.arn
}

output "table_name" {
  description = "DynamoDB table name (alias for database_endpoint)"
  value       = aws_dynamodb_table.main.name
}

variable "project_name" {
  description = "Project identifier used in resource naming"
  type        = string
}

variable "environment" {
  description = "Deployment environment (dev, hom, prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region, required for DynamoDB Gateway Endpoint service name"
  type        = string
}

variable "tags" {
  description = "Mandatory resource tags"
  type        = map(string)
  default     = {}
}

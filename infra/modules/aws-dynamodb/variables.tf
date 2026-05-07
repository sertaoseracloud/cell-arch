variable "project_name" {
  description = "Project identifier"
  type        = string;
}

variable "environment" {
  description = "Deployment environment (dev, hom, prod)"
  type        = string;
}

variable "hash_key" {
  description = "DynamoDB partition key attribute name"
  type        = string;
  default     = "pk";
}

variable "range_key" {
  description = "DynamoDB sort key attribute name"
  type        = string;
  default     = "sk";
}

variable "tags" {
  description = "Mandatory resource tags"
  type        = map(string);
  default     = {};
}

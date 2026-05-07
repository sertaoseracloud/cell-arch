variable "project_name" {
  description = "Project identifier"
  type        = string
}

variable "environment" {
  description = "Deployment environment (dev, hom, prod)"
  type        = string
}

variable "tags" {
  description = "Mandatory resource tags"
  type        = map(string)
  default     = {}
}

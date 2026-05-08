variable "project_name" {
  type    = string
  default = "cell-arch"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "tags" {
  type    = map(string)
  default = {
    Project     = "cell-arch"
    Environment = "dev"
    ManagedBy  = "terraform"
  }
}

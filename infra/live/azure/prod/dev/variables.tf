variable "project_name" {
  type    = string
  default = "cell-arch"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "location" {
  type    = string
  default = "eastus"
}

variable "tags" {
  type    = map(string)
  default = {
    Project     = "cell-arch"
    Environment = "dev"
    ManagedBy  = "terraform"
  }
}

terraform {
  required_version = "~> 1.7.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.101.0"
    }
  }

  backend "azurerm" {
    resource_group_name  = "cell-arch-bootstrap-dev"
    storage_account_name = "cellarchtfstatedev"
    container_name       = "tfstate"
    key                  = "dev/terraform.tfstate"
  }
}

provider "azurerm" {
  features {}
}

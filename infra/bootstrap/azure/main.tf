terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.101.0"
    }
  }
}

variable "project_name" {
  description = "Project identifier"
  type        = string
}

variable "environment" {
  description = "Environment (dev, hom, prod)"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}

# ── Resource Group for bootstrap resources ─────────────────────────

resource "azurerm_resource_group" "bootstrap" {
  name     = "${var.project_name}-bootstrap-${var.environment}"
  location = var.location
  tags     = var.tags
}

# ── Storage Account + Container for Terraform state ───────────────

resource "azurerm_storage_account" "tfstate" {
  name                     = "${lower(var.project_name)}tfstate${lower(var.environment)}"
  resource_group_name      = azurerm_resource_group.bootstrap.name
  location                 = azurerm_resource_group.bootstrap.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  tags                     = var.tags
}

resource "azurerm_storage_container" "tfstate" {
  name                  = "tfstate"
  storage_account_name = azurerm_storage_account.tfstate.name
}

# ── Lease for state locking (simulates DynamoDB lock) ───────

# Note: Azure Blob native leasing provides locking; no separate resource needed.
# The lease is acquired via Terraform's azurerm backend `lease` feature.

output "storage_account_name" {
  description = "Storage account name for Terraform state"
  value       = azurerm_storage_account.tfstate.name
}

output "container_name" {
  description = "Blob container for Terraform state"
  value       = azurerm_storage_container.tfstate.name
}

output "resource_group_name" {
  description = "Resource group containing bootstrap resources"
  value       = azurerm_resource_group.bootstrap.name
}

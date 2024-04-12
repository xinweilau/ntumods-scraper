terraform {
  required_providers {
    azurerm = {
      source = "hashicorp/azurerm"
      version = "~> 3.98.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "rg" {
  name     = "cz4052"
  location = "southeastasia"
}

resource "azurerm_storage_account" "sa" {
  name                              = "ntumodssa"
  resource_group_name               = azurerm_resource_group.rg.name
  location                          = azurerm_resource_group.rg.location
  account_tier                      = "Standard"
  account_replication_type          = "LRS"
  public_network_access_enabled     = true
  allow_nested_items_to_be_public   = true

  blob_properties {
    cors_rule {
      allowed_headers = ["*"]
      allowed_methods = ["GET", "OPTIONS"]
      allowed_origins = ["*"]
      exposed_headers = ["*"]
      max_age_in_seconds = 86400
    }
  }
}

resource "azurerm_storage_container" "sc" {
  name                  = "ntumodssc"
  storage_account_name  = azurerm_storage_account.sa.name
  container_access_type = "blob"
}

resource "azurerm_storage_management_policy" "storage_policy" {
  storage_account_id = azurerm_storage_account.sa.id

  rule {
    name    = "change_blob_tier_after_30_days"
    enabled = true

    filters {
      prefix_match = ["${azurerm_storage_container.sc.name}/"]
      blob_types   = ["blockBlob"]
    }

    actions {
      base_blob {
        tier_to_archive_after_days_since_creation_greater_than = 365
        tier_to_cool_after_days_since_creation_greater_than    = 30
      }
    }
  }
}

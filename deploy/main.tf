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
}

resource "azurerm_storage_container" "sc" {
  name                  = "ntumodssc"
  storage_account_name  = azurerm_storage_account.sa.name
  container_access_type = "blob"
}

resource "azurerm_storage_management_policy" "example" {
  storage_account_id = azurerm_storage_account.sa.id

  rule {
    name    = "rule1"
    enabled = true

    filters {
      prefix_match = ["ntumodssc/"]
      blob_types   = ["blockBlob"]
    }
    
    actions {
      base_blob {
        tier_to_cool_after_days_since_last_access_time_greater_than = 30
      }
    }
  }
}

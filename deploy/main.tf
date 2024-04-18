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

# Host the NTUMods Scraper
variable "docker_hub_username" {}
variable "docker_hub_password" {}

resource "azurerm_container_group" "acg" {
  name                = "ntumods-scraper"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  os_type             = "Linux"
  restart_policy      = "OnFailure"
  dns_name_label      = "ntumods-scraper"

  container {
    name   = "ntumods-scraper"
    image  = "xinweilau/myrepopo:ntumods-scraper"
    cpu    = 0.25
    memory = 0.5

    ports {
      port     = 8080
      protocol = "TCP"
    }
  }

  image_registry_credential {
    server   = "index.docker.io"
    username = var.docker_hub_username
    password = var.docker_hub_password
  }
}

resource "azurerm_logic_app_workflow" "lapp" {
  name                = "ntumodsscraper"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
}

resource "azurerm_logic_app_trigger_http_request" "ltrig_req" {
  name         = "trigger-scraper-req"
  logic_app_id = azurerm_logic_app_workflow.lapp.id

  schema = <<SCHEMA
  {
      "type": "object",
      "properties": {
          "method": {
              "type": "string"
          }
      }
  }
  SCHEMA
}

resource "azurerm_logic_app_action_http" "lact" {
  name         = "request-to-scrape"
  logic_app_id = azurerm_logic_app_workflow.lapp.id

  method  = "GET"
  uri     = "http://${azurerm_container_group.acg.dns_name_label}.southeastasia.azurecontainer.io:8080"
}

resource "azurerm_logic_app_trigger_recurrence" "sem1" {
  name         = "run-for-semester1"
  logic_app_id = azurerm_logic_app_workflow.lapp.id
  frequency    = "Month"
  start_time   = "2024-06-01T00:00:00Z"
  time_zone    = "Singapore Standard Time"
  interval     = 12
}

resource "azurerm_logic_app_trigger_recurrence" "sem2" {
  name         = "run-for-semester2"
  logic_app_id = azurerm_logic_app_workflow.lapp.id
  frequency    = "Month"
  start_time   = "2024-12-01T00:00:00Z"
  time_zone    = "Singapore Standard Time"
  interval     = 12
}

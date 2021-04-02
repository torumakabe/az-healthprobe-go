terraform {
  required_version = "~> 0.14.9"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 2.53"
    }

    github = {
      source  = "integrations/github"
      version = "~> 4.6"
    }

  }
}

provider "azurerm" {
  features {}
}

data "azurerm_monitor_action_group" "email" {
  name                = var.alert_actiongroup
  resource_group_name = var.alert_actiongroup_rg
}

resource "azurerm_resource_group" "healthprobe" {
  name     = var.healthprobe_rg
  location = var.healthprobe_location
}

resource "azurerm_application_insights" "healthprobe" {
  name                = var.app_insights
  resource_group_name = azurerm_resource_group.healthprobe.name
  location            = azurerm_resource_group.healthprobe.location
  application_type    = "other"
}

resource "azurerm_virtual_network" "default" {
  name                = "vnet-default"
  resource_group_name = azurerm_resource_group.healthprobe.name
  location            = azurerm_resource_group.healthprobe.location
  address_space       = ["10.0.0.0/8"]
}

resource "azurerm_subnet" "probe" {
  name                 = "subnet-probe"
  resource_group_name  = azurerm_resource_group.healthprobe.name
  virtual_network_name = azurerm_virtual_network.default.name
  address_prefixes     = ["10.0.1.0/27"]

  delegation {
    name = "delegation"

    service_delegation {
      name    = "Microsoft.ContainerInstance/containerGroups"
      actions = ["Microsoft.Network/virtualNetworks/subnets/action"]
    }
  }
}

resource "azurerm_network_profile" "probe" {
  name                = "nwprof-probe"
  resource_group_name = azurerm_resource_group.healthprobe.name
  location            = azurerm_resource_group.healthprobe.location

  container_network_interface {
    name = "nic-probe"

    ip_configuration {
      name      = "ipconf-probe"
      subnet_id = azurerm_subnet.probe.id
    }
  }
}

resource "azurerm_monitor_metric_alert" "target_availability" {
  name                = "alert-target-avail"
  resource_group_name = azurerm_resource_group.healthprobe.name
  scopes              = [azurerm_application_insights.healthprobe.id]
  description         = "Alert when site availability has fallen below the threshold"
  frequency           = "PT1M"
  window_size         = "PT1H"
  severity            = 2
  enabled             = false

  criteria {
    metric_namespace = "Microsoft.Insights/components"
    metric_name      = "availabilityResults/availabilityPercentage"
    aggregation      = "Average"
    operator         = "LessThan"
    threshold        = 95

    dimension {
      name     = "availabilityResult/name"
      operator = "Include"
      values   = ["*"]
    }
  }

  action {
    action_group_id = data.azurerm_monitor_action_group.email.id
  }
}

resource "azurerm_monitor_metric_alert" "probe_availability" {
  name                = "alert-probe-avail"
  resource_group_name = azurerm_resource_group.healthprobe.name
  scopes              = [azurerm_application_insights.healthprobe.id]
  description         = "Alert when cannot receive telemetry from probe"
  frequency           = "PT1M"
  window_size         = "PT30M"
  severity            = 3
  enabled             = false

  criteria {
    metric_namespace = "Microsoft.Insights/components"
    metric_name      = "availabilityResults/count"
    aggregation      = "Count"
    operator         = "LessThan"
    threshold        = 1
  }

  action {
    action_group_id = data.azurerm_monitor_action_group.email.id
  }
}

provider "github" {
  token = var.github_token
}

data "github_actions_public_key" "healthprobe" {
  repository = var.github_repo
}

resource "github_actions_secret" "gh_pat" {
  repository      = var.github_repo
  secret_name     = "GH_PAT"
  plaintext_value = var.github_token
}

resource "github_actions_secret" "azure_credentials" {
  repository      = var.github_repo
  secret_name     = "AZURE_CREDENTIALS"
  plaintext_value = var.azure_credentials
}

resource "github_actions_secret" "resource_group" {
  repository      = var.github_repo
  secret_name     = "RESOURCE_GROUP"
  plaintext_value = var.healthprobe_rg
}

resource "github_actions_secret" "probe_location" {
  repository      = var.github_repo
  secret_name     = "PROBE_LOCATION"
  plaintext_value = var.healthprobe_location
}

resource "github_actions_secret" "probe_ikey" {
  repository      = var.github_repo
  secret_name     = "PROBE_INSTRUMENTATION_KEY"
  plaintext_value = azurerm_application_insights.healthprobe.instrumentation_key
}

resource "github_actions_secret" "registry_login_server" {
  repository      = var.github_repo
  secret_name     = "REGISTRY_LOGIN_SERVER"
  plaintext_value = var.registry_login_server
}

resource "github_actions_secret" "registry_login_username" {
  repository      = var.github_repo
  secret_name     = "REGISTRY_USERNAME"
  plaintext_value = var.registry_username
}

resource "github_actions_secret" "registry_password" {
  repository      = var.github_repo
  secret_name     = "REGISTRY_PASSWORD"
  plaintext_value = var.registry_password
}

resource "github_actions_secret" "probe_nw_profile" {
  repository      = var.github_repo
  secret_name     = "PROBE_NW_PROFILE"
  plaintext_value = azurerm_network_profile.probe.id
}

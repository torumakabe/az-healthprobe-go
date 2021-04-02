variable "healthprobe_rg" {
  type = string
}

variable "healthprobe_location" {
  type = string
}

variable "app_insights" {
  type = string
}

variable "alert_actiongroup" {
  type = string
}

variable "alert_actiongroup_rg" {
  type = string
}

variable "github_repo" {
  type = string
}

variable "github_token" {
  type      = string
  sensitive = true
}

variable "azure_credentials" {
  type      = string
  sensitive = true
}

variable "registry_login_server" {
  type      = string
  sensitive = true
}

variable "registry_username" {
  type      = string
  sensitive = true
}

variable "registry_password" {
  type      = string
  sensitive = true
}

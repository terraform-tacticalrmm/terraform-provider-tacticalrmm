# examples/script_export_example.tf
terraform {
  required_providers {
    tacticalrmm = {
      source  = "terraform-tacticalrmm/tacticalrmm"
      version = "~> 0.1.0"
    }
  }
}

provider "tacticalrmm" {
  endpoint = var.trmm_endpoint
  api_key  = var.trmm_api_key
}

# Fetch all scripts from TRMM
data "tacticalrmm_scripts" "all_scripts" {
  # No filters - export all scripts
}

# Export scripts to local filesystem
module "script_export" {
  source = "../modules/script_exporter"
  
  scripts          = data.tacticalrmm_scripts.all_scripts.scripts
  output_directory = "${path.module}/exported_scripts"
  include_manifest = true
}

# Display export results
output "export_summary" {
  description = "Summary of exported scripts"
  value       = module.script_export.export_statistics
}

output "directory_structure" {
  description = "Generated directory structure"
  value       = module.script_export.directory_structure
}

# Example with filtering
data "tacticalrmm_scripts" "user_scripts_only" {
  script_type = "userdefined"
}

module "user_script_export" {
  source = "../modules/script_exporter"
  
  scripts          = data.tacticalrmm_scripts.user_scripts_only.scripts
  output_directory = "${path.module}/user_scripts"
  include_manifest = true
}

# Variables
variable "trmm_endpoint" {
  description = "Tactical RMM API endpoint"
  type        = string
}

variable "trmm_api_key" {
  description = "Tactical RMM API key"
  type        = string
  sensitive   = true
}

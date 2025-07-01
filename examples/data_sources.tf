terraform {
  required_providers {
    tacticalrmm = {
      source = "terraform-tacticalrmm/tacticalrmm"
      version = "~> 0.1.0"
    }
  }
}

provider "tacticalrmm" {
  endpoint = "https://api.your-trmm-instance.com"
  api_key  = var.trmm_api_key
}

# Example 1: Look up a script by ID
data "tacticalrmm_script" "by_id" {
  id = 42
}

# Example 2: Look up a script by name
data "tacticalrmm_script" "disk_cleanup" {
  name = "Disk Cleanup Script"
}

# Example 3: Look up a script snippet by name
data "tacticalrmm_script_snippet" "common_functions" {
  name = "GetDiskSpace"
}

# Example 4: Look up a keystore entry by name
data "tacticalrmm_keystore" "api_endpoint" {
  name = "external_api_endpoint"
}

# Example 5: Use data sources in resources
resource "tacticalrmm_script" "monitoring_script" {
  name        = "Advanced Monitoring"
  description = "Monitoring script that uses existing components"
  shell       = "powershell"
  category    = "Monitoring"
  
  # Reference the script body from an existing script
  script_body = <<-EOT
    # Include common functions from snippet
    ${data.tacticalrmm_script_snippet.common_functions.code}
    
    # Get API endpoint from keystore
    $apiEndpoint = "${data.tacticalrmm_keystore.api_endpoint.value}"
    
    # Your custom monitoring logic here
    Write-Output "Using API endpoint: $apiEndpoint"
    
    # Run disk space check
    {{GetDiskSpace}}
  EOT
  
  default_timeout = 60
}

# Example 6: Output data source values
output "script_details" {
  value = {
    id          = data.tacticalrmm_script.disk_cleanup.id
    name        = data.tacticalrmm_script.disk_cleanup.name
    description = data.tacticalrmm_script.disk_cleanup.description
    shell       = data.tacticalrmm_script.disk_cleanup.shell
    timeout     = data.tacticalrmm_script.disk_cleanup.default_timeout
  }
  description = "Details of the disk cleanup script"
}

output "snippet_code" {
  value       = data.tacticalrmm_script_snippet.common_functions.code
  description = "Code from the common functions snippet"
  sensitive   = false
}

# Note: keystore value is sensitive, so it won't be displayed in plain text
output "keystore_value" {
  value       = data.tacticalrmm_keystore.api_endpoint.value
  description = "Value from keystore (sensitive)"
  sensitive   = true
}

# Variables
variable "trmm_api_key" {
  description = "Tactical RMM API Key"
  type        = string
  sensitive   = true
}

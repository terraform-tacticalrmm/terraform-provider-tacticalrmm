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

# ===== PLURAL DATA SOURCES (fetch multiple items) =====

# Example 1: Fetch ALL scripts
data "tacticalrmm_scripts" "all_scripts" {
  # No filters - fetches all scripts
}

# Example 2: Fetch ALL script snippets
data "tacticalrmm_script_snippets" "all_snippets" {
  # No filters - fetches all snippets
}

# Example 3: Fetch ALL keystore entries
data "tacticalrmm_keystores" "all_keystores" {
  # No filters - fetches all keystore entries
}

# Example 4: Filter scripts by ID using plural data source
data "tacticalrmm_scripts" "filtered_by_id" {
  id = 42
}

# Example 5: Filter scripts by name using plural data source
data "tacticalrmm_scripts" "filtered_by_name" {
  name = "Disk Cleanup Script"
}

# ===== SINGULAR DATA SOURCES (fetch single item) =====

# Example 6: Look up a single script by ID
data "tacticalrmm_script" "single_by_id" {
  id = 42
}

# Example 7: Look up a single script by name
data "tacticalrmm_script" "disk_cleanup" {
  name = "Disk Cleanup Script"
}

# Example 8: Look up a single script snippet by name
data "tacticalrmm_script_snippet" "common_functions" {
  name = "GetDiskSpace"
}

# Example 9: Look up a single keystore entry by name
data "tacticalrmm_keystore" "api_endpoint" {
  name = "external_api_endpoint"
}

# ===== USING DATA SOURCES =====

# Example 10: Use scripts data to create dynamic resources
resource "tacticalrmm_script" "backup_scripts" {
  for_each = { for script in data.tacticalrmm_scripts.all_scripts.scripts : script.name => script if script.category == "Backup" }
  
  name        = "${each.value.name}-copy"
  description = "Copy of ${each.value.name}"
  shell       = each.value.shell
  category    = "Backup-Copies"
  script_body = each.value.script_body
  
  default_timeout = each.value.default_timeout
}

# Example 11: Create a report of all PowerShell scripts
locals {
  powershell_scripts = [
    for script in data.tacticalrmm_scripts.all_scripts.scripts : {
      name        = script.name
      description = script.description
      timeout     = script.default_timeout
    } if script.shell == "powershell"
  ]
}

# Example 12: Use all snippets to build a reference
output "all_snippet_names" {
  value = [for snippet in data.tacticalrmm_script_snippets.all_snippets.snippets : snippet.name]
  description = "List of all available script snippet names"
}

# Example 13: Count scripts by category
output "scripts_by_category" {
  value = {
    for category in distinct([for s in data.tacticalrmm_scripts.all_scripts.scripts : coalesce(s.category, "uncategorized")]) :
    category => length([for s in data.tacticalrmm_scripts.all_scripts.scripts : s if coalesce(s.category, "uncategorized") == category])
  }
  description = "Count of scripts grouped by category"
}

# Example 14: Find all keystore entries that look like API keys
output "api_related_keys" {
  value = [
    for key in data.tacticalrmm_keystores.all_keystores.keystores : key.name
    if can(regex("(?i)(api|key|token|secret)", key.name))
  ]
  description = "Keystore entries that might be API-related"
}

# Example 15: Use filtered results
output "filtered_script_count" {
  value = length(data.tacticalrmm_scripts.filtered_by_name.scripts)
  description = "Number of scripts matching the name filter"
}

# Example 16: Create a consolidated script using multiple snippets
resource "tacticalrmm_script" "master_script" {
  name        = "Master Utility Script"
  description = "Combines multiple snippets"
  shell       = "powershell"
  category    = "Utilities"
  
  script_body = <<-EOT
    # Master script combining all utility snippets
    
    ${join("\n\n", [
      for snippet in data.tacticalrmm_script_snippets.all_snippets.snippets : 
      "# === ${snippet.name} ===\n${snippet.code}"
      if snippet.shell == "powershell" && can(regex("(?i)utility", snippet.name))
    ])}
    
    # Main execution
    Write-Output "Utility functions loaded"
  EOT
  
  default_timeout = 300
}

# Variables
variable "trmm_api_key" {
  description = "Tactical RMM API Key"
  type        = string
  sensitive   = true
}

# trmm_script Data Source

## Overview

The `trmm_script` data source enables querying individual automation scripts within Tactical RMM by ID or name, facilitating script reference and configuration inheritance patterns.

## Technical Specifications

### Data Source Schema

```hcl
data "trmm_script" "example" {
  # Query Parameters (mutually exclusive)
  id   = number
  name = string
  
  # Computed Attributes
  description          = string
  shell               = string
  script_type         = string
  category            = string
  script_body         = string
  default_timeout     = number
  favorite            = bool
  hidden              = bool
  run_as_user         = bool
  args                = list(string)
  env_vars            = list(string)
  supported_platforms = list(string)
  syntax              = string
}
```

### Query Parameters

| Parameter | Type | Description | Priority |
|-----------|------|-------------|----------|
| `id` | Number | Script identifier | Primary |
| `name` | String | Script name (exact match) | Secondary |

**Note**: Provide either `id` or `name`, not both. ID takes precedence if both supplied.

### Computed Attributes

All script attributes are exposed as computed values matching the resource schema.

## Implementation Patterns

### Pattern 1: Reference Existing Scripts

```hcl
# Query by ID for precise reference
data "trmm_script" "maintenance" {
  id = 123
}

# Query by name for dynamic lookup
data "trmm_script" "by_name" {
  name = "System Maintenance Script"
}

# Reference in other resources
resource "trmm_task" "scheduled_maintenance" {
  name        = "Weekly Maintenance"
  script_id   = data.trmm_script.maintenance.id
  schedule    = "0 2 * * 0"  # Sunday 2 AM
  enabled     = true
}
```

### Pattern 2: Script Inheritance and Extension

```hcl
# Base script lookup
data "trmm_script" "base_monitor" {
  name = "Base System Monitor"
}

# Extended script with additional functionality
resource "trmm_script" "enhanced_monitor" {
  name        = "Enhanced ${data.trmm_script.base_monitor.name}"
  description = "Extended monitoring with alerting"
  shell       = data.trmm_script.base_monitor.shell
  category    = data.trmm_script.base_monitor.category
  
  script_body = <<-EOT
    # Include base monitoring logic
    ${data.trmm_script.base_monitor.script_body}
    
    # Additional enhanced monitoring
    function Send-Alert {
      param($Message, $Severity)
      # Alert implementation
    }
    
    # Enhanced threshold checking
    if ($CPUUsage -gt 90) {
      Send-Alert -Message "Critical CPU usage: $CPUUsage%" -Severity "Critical"
    }
  EOT
  
  # Inherit configuration with modifications
  default_timeout     = data.trmm_script.base_monitor.default_timeout * 2
  supported_platforms = data.trmm_script.base_monitor.supported_platforms
  
  env_vars = concat(
    data.trmm_script.base_monitor.env_vars,
    ["ALERT_ENABLED=true", "ALERT_THRESHOLD=90"]
  )
}
```

### Pattern 3: Script Validation and Compliance

```hcl
# Define compliance requirements
locals {
  required_scripts = {
    security_baseline = {
      name     = "Security Baseline Check"
      category = "Security"
      timeout  = 300
    }
    patch_assessment = {
      name     = "Patch Assessment"
      category = "Maintenance"
      timeout  = 600
    }
    backup_verification = {
      name     = "Backup Verification"
      category = "Backup"
      timeout  = 900
    }
  }
}

# Verify required scripts exist
data "trmm_script" "compliance_check" {
  for_each = local.required_scripts
  name     = each.value.name
}

# Validate script configurations
resource "null_resource" "script_compliance" {
  for_each = local.required_scripts
  
  lifecycle {
    precondition {
      condition = (
        data.trmm_script.compliance_check[each.key].category == each.value.category &&
        data.trmm_script.compliance_check[each.key].default_timeout >= each.value.timeout
      )
      error_message = "Script ${each.value.name} does not meet compliance requirements"
    }
  }
}
```

### Pattern 4: Dynamic Script Selection

```hcl
# Environment-based script selection
variable "environment" {
  description = "Deployment environment"
  type        = string
  validation {
    condition     = contains(["production", "staging", "development"], var.environment)
    error_message = "Environment must be production, staging, or development"
  }
}

locals {
  script_mapping = {
    production = {
      monitoring = "Production System Monitor"
      backup     = "Production Backup Script"
      cleanup    = "Production Cleanup Script"
    }
    staging = {
      monitoring = "Staging System Monitor"
      backup     = "Staging Backup Script"
      cleanup    = "Staging Cleanup Script"
    }
    development = {
      monitoring = "Dev System Monitor"
      backup     = "Dev Backup Script"
      cleanup    = "Dev Cleanup Script"
    }
  }
}

# Dynamic script lookup based on environment
data "trmm_script" "env_scripts" {
  for_each = local.script_mapping[var.environment]
  name     = each.value
}

# Use environment-specific scripts
resource "trmm_policy" "env_policy" {
  name        = "${var.environment} Standard Policy"
  description = "Standard policy for ${var.environment} environment"
  
  scripts = [
    for key, script in data.trmm_script.env_scripts : {
      script_id = script.id
      schedule  = key == "monitoring" ? "*/15 * * * *" : "0 3 * * *"
    }
  ]
}
```

## Advanced Usage Patterns

### Pattern 1: Script Dependency Management

```hcl
# Define script dependencies
locals {
  script_dependencies = {
    "Application Deployment" = ["Environment Setup", "Prerequisite Check"]
    "Database Maintenance"   = ["Database Backup", "Service Stop"]
    "Security Hardening"     = ["Baseline Assessment", "Configuration Backup"]
  }
}

# Retrieve all scripts with dependencies
data "trmm_script" "all_scripts" {
  for_each = toset(flatten([
    for script, deps in local.script_dependencies : concat([script], deps)
  ]))
  
  name = each.value
}

# Validate dependencies exist
resource "null_resource" "dependency_validation" {
  for_each = local.script_dependencies
  
  lifecycle {
    precondition {
      condition = alltrue([
        for dep in each.value : contains(
          [for s in data.trmm_script.all_scripts : s.name],
          dep
        )
      ])
      error_message = "Missing dependency for ${each.key}"
    }
  }
}
```

### Pattern 2: Script Metadata Extraction

```hcl
# Extract and process script metadata
data "trmm_script" "analyzed_scripts" {
  for_each = toset(var.script_names)
  name     = each.value
}

# Generate script analysis report
locals {
  script_analysis = {
    for name, script in data.trmm_script.analyzed_scripts : name => {
      complexity = length(split("\n", script.script_body))
      has_error_handling = contains(
        lower(script.script_body),
        script.shell == "powershell" ? "try" : "trap"
      )
      uses_env_vars = length(coalesce(script.env_vars, [])) > 0
      platform_count = length(coalesce(script.supported_platforms, ["all"]))
      estimated_runtime = script.default_timeout
    }
  }
  
  # Summary statistics
  script_summary = {
    total_scripts = length(local.script_analysis)
    average_complexity = avg([for s in values(local.script_analysis) : s.complexity])
    scripts_with_error_handling = length([
      for s in values(local.script_analysis) : s if s.has_error_handling
    ])
  }
}

output "script_analysis_report" {
  value = {
    details = local.script_analysis
    summary = local.script_summary
  }
}
```

### Pattern 3: Cross-Reference Validation

```hcl
# Validate script references in configuration
variable "task_configurations" {
  type = list(object({
    name      = string
    script_id = number
    schedule  = string
  }))
}

# Retrieve all referenced scripts
data "trmm_script" "referenced" {
  for_each = {
    for task in var.task_configurations : 
    tostring(task.script_id) => task.script_id
  }
  
  id = each.value
}

# Validate all references resolve
resource "null_resource" "reference_validation" {
  lifecycle {
    precondition {
      condition = alltrue([
        for task in var.task_configurations :
        can(data.trmm_script.referenced[tostring(task.script_id)])
      ])
      error_message = "Invalid script reference in task configuration"
    }
  }
}
```

## Error Handling

### Common Query Errors

1. **Script Not Found**
   ```
   Error: Script with name "NonExistent Script" not found
   ```
   - Verify exact script name match
   - Check for typos or case sensitivity
   - Ensure script exists in target environment

2. **Ambiguous Query**
   ```
   Error: Multiple scripts found with similar names
   ```
   - Use ID for precise lookup
   - Ensure unique script names
   - Implement naming conventions

3. **Permission Denied**
   ```
   Error: Insufficient permissions to read script
   ```
   - Verify API key permissions
   - Check script access policies
   - Review organizational restrictions

### Diagnostic Queries

```hcl
# Diagnostic data source for troubleshooting
data "trmm_script" "diagnostic" {
  count = var.enable_diagnostics ? 1 : 0
  id    = var.diagnostic_script_id
}

output "diagnostic_script_info" {
  value = var.enable_diagnostics ? {
    found       = data.trmm_script.diagnostic[0].id != null
    name        = data.trmm_script.diagnostic[0].name
    category    = data.trmm_script.diagnostic[0].category
    shell       = data.trmm_script.diagnostic[0].shell
    timeout     = data.trmm_script.diagnostic[0].default_timeout
    body_length = length(data.trmm_script.diagnostic[0].script_body)
  } : null
  
  sensitive = true
}
```

## Best Practices

### 1. Query Optimization
- Use ID queries when possible for performance
- Cache results in locals for multiple references
- Implement error handling for missing scripts

### 2. Security Considerations
- Mark sensitive script content appropriately
- Validate script sources before execution
- Implement least-privilege access patterns

### 3. Maintainability
- Document script dependencies clearly
- Use consistent naming conventions
- Implement version tracking mechanisms

## Related Resources

- [trmm_scripts](scripts.md) - List multiple scripts
- [Resource: trmm_script](../resources/script.md) - Manage scripts
- [trmm_script_snippet](script_snippet.md) - Query script snippets

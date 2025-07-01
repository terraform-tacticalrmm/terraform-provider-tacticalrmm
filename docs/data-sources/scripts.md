# trmm_scripts Data Source

## Overview

The `trmm_scripts` data source provides comprehensive querying capabilities for multiple automation scripts within Tactical RMM, enabling filtered retrieval and bulk operations with systematic precision.

## Technical Specifications

### Data Source Schema

```hcl
data "trmm_scripts" "example" {
  # Filter Parameters (all optional)
  category            = string
  shell              = string
  script_type        = string
  hidden             = bool
  favorite           = bool
  supported_platform = string
  
  # Computed Results
  scripts = list(object({
    id                  = number
    name                = string
    description         = string
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
  }))
}
```

### Filter Parameters

| Parameter | Type | Description | Filter Logic |
|-----------|------|-------------|--------------|
| `category` | String | Script category filter | Exact match |
| `shell` | String | Execution environment filter | Exact match: `powershell`, `cmd`, `python`, `shell` |
| `script_type` | String | Script type filter | `userdefined` or `builtin` |
| `hidden` | Bool | Hidden status filter | Include/exclude hidden scripts |
| `favorite` | Bool | Favorite status filter | Include only favorites |
| `supported_platform` | String | Platform compatibility filter | Contains match: `windows`, `linux`, `darwin` |

## Implementation Patterns

### Pattern 1: Category-Based Script Retrieval

```hcl
# Retrieve all maintenance scripts
data "trmm_scripts" "maintenance_scripts" {
  category = "Maintenance"
}

# Retrieve all security scripts
data "trmm_scripts" "security_scripts" {
  category = "Security"
}

# Generate task schedules for maintenance scripts
resource "trmm_task" "scheduled_maintenance" {
  for_each = {
    for script in data.trmm_scripts.maintenance_scripts.scripts :
    script.name => script if script.hidden == false
  }
  
  name        = "Scheduled: ${each.value.name}"
  script_id   = each.value.id
  schedule    = "0 3 * * *"  # Daily at 3 AM
  enabled     = true
  timeout     = each.value.default_timeout
}
```

### Pattern 2: Platform-Specific Script Management

```hcl
# Platform-specific script retrieval
data "trmm_scripts" "windows_scripts" {
  supported_platform = "windows"
  script_type       = "userdefined"
}

data "trmm_scripts" "linux_scripts" {
  supported_platform = "linux"
  script_type       = "userdefined"
}

data "trmm_scripts" "cross_platform" {
  shell = "python"  # Python scripts typically cross-platform
}

# Generate platform-specific policies
locals {
  platform_policies = {
    windows = {
      name    = "Windows Automation Policy"
      scripts = data.trmm_scripts.windows_scripts.scripts
    }
    linux = {
      name    = "Linux Automation Policy"
      scripts = data.trmm_scripts.linux_scripts.scripts
    }
  }
}

resource "trmm_policy" "platform_specific" {
  for_each = local.platform_policies
  
  name        = each.value.name
  description = "Automated policy for ${each.key} systems"
  
  dynamic "script_assignment" {
    for_each = each.value.scripts
    content {
      script_id = script_assignment.value.id
      order     = script_assignment.key
    }
  }
}
```

### Pattern 3: Script Inventory Analysis

```hcl
# Comprehensive script inventory
data "trmm_scripts" "all_scripts" {}

# Analyze script distribution
locals {
  script_analysis = {
    # Category distribution
    by_category = {
      for cat in distinct([for s in data.trmm_scripts.all_scripts.scripts : s.category]) :
      cat => length([
        for s in data.trmm_scripts.all_scripts.scripts : s
        if s.category == cat
      ])
    }
    
    # Shell type distribution
    by_shell = {
      for shell in distinct([for s in data.trmm_scripts.all_scripts.scripts : s.shell]) :
      shell => length([
        for s in data.trmm_scripts.all_scripts.scripts : s
        if s.shell == shell
      ])
    }
    
    # Platform coverage
    platform_coverage = {
      windows_only = length([
        for s in data.trmm_scripts.all_scripts.scripts : s
        if length(s.supported_platforms) == 1 && contains(s.supported_platforms, "windows")
      ])
      linux_only = length([
        for s in data.trmm_scripts.all_scripts.scripts : s
        if length(s.supported_platforms) == 1 && contains(s.supported_platforms, "linux")
      ])
      cross_platform = length([
        for s in data.trmm_scripts.all_scripts.scripts : s
        if length(s.supported_platforms) > 1
      ])
    }
    
    # Complexity metrics
    complexity = {
      total_scripts = length(data.trmm_scripts.all_scripts.scripts)
      with_args     = length([for s in data.trmm_scripts.all_scripts.scripts : s if length(coalesce(s.args, [])) > 0])
      with_env_vars = length([for s in data.trmm_scripts.all_scripts.scripts : s if length(coalesce(s.env_vars, [])) > 0])
      long_running  = length([for s in data.trmm_scripts.all_scripts.scripts : s if s.default_timeout > 300])
    }
  }
}

output "script_inventory_report" {
  value = local.script_analysis
}
```

### Pattern 4: Favorite Scripts Dashboard

```hcl
# Retrieve favorite scripts for quick access
data "trmm_scripts" "favorites" {
  favorite = true
  hidden   = false
}

# Create quick-access tasks for favorites
resource "trmm_task" "favorite_quick_access" {
  for_each = {
    for script in data.trmm_scripts.favorites.scripts :
    script.name => script
  }
  
  name        = "Quick: ${each.value.name}"
  script_id   = each.value.id
  enabled     = false  # Manual trigger only
  
  tags = ["favorite", "quick-access", lower(each.value.category)]
}

# Generate favorite scripts documentation
resource "local_file" "favorites_documentation" {
  filename = "${path.module}/docs/favorite-scripts.md"
  content  = <<-EOT
    # Favorite Scripts Reference
    
    Generated: ${timestamp()}
    
    ## Scripts by Category
    
    %{for category in distinct([for s in data.trmm_scripts.favorites.scripts : s.category])}
    ### ${category}
    
    %{for script in [for s in data.trmm_scripts.favorites.scripts : s if s.category == category]}
    #### ${script.name}
    - **ID**: ${script.id}
    - **Shell**: ${script.shell}
    - **Timeout**: ${script.default_timeout}s
    - **Platforms**: ${join(", ", coalesce(script.supported_platforms, ["all"]))}
    - **Description**: ${coalesce(script.description, "No description")}
    
    %{endfor}
    %{endfor}
  EOT
}
```

## Advanced Usage Patterns

### Pattern 1: Script Dependency Resolution

```hcl
# Define script dependency graph
locals {
  script_dependencies = {
    "System Update" = {
      depends_on = ["System Backup", "Service Stop"]
      category   = "Maintenance"
    }
    "Security Scan" = {
      depends_on = ["Update Definitions", "System Baseline"]
      category   = "Security"
    }
  }
}

# Retrieve all scripts to validate dependencies
data "trmm_scripts" "dependency_check" {}

# Build dependency validation map
locals {
  available_scripts = toset([for s in data.trmm_scripts.dependency_check.scripts : s.name])
  
  dependency_validation = {
    for script, config in local.script_dependencies : script => {
      exists = contains(local.available_scripts, script)
      dependencies_met = alltrue([
        for dep in config.depends_on : contains(local.available_scripts, dep)
      ])
      missing_deps = [
        for dep in config.depends_on : dep
        if !contains(local.available_scripts, dep)
      ]
    }
  }
}

# Validate all dependencies are satisfied
resource "null_resource" "dependency_validator" {
  for_each = local.dependency_validation
  
  lifecycle {
    precondition {
      condition     = each.value.exists && each.value.dependencies_met
      error_message = "Script '${each.key}' dependencies not satisfied: ${join(", ", each.value.missing_deps)}"
    }
  }
}
```

### Pattern 2: Script Version Tracking

```hcl
# Track script modifications using checksums
data "trmm_scripts" "versioned" {}

# Generate version tracking metadata
locals {
  script_versions = {
    for script in data.trmm_scripts.versioned.scripts : script.name => {
      id       = script.id
      checksum = md5(script.script_body)
      size     = length(script.script_body)
      metadata = {
        category        = script.category
        shell          = script.shell
        last_tracked   = timestamp()
        complexity_score = length(split("\n", script.script_body))
      }
    }
  }
}

# Store version tracking in state
resource "terraform_data" "script_versions" {
  input = local.script_versions
  
  lifecycle {
    create_before_destroy = true
  }
}

# Detect changes between runs
output "script_changes" {
  value = try(
    {
      for name, current in local.script_versions :
      name => {
        changed = current.checksum != try(
          terraform_data.script_versions.output[name].checksum,
          ""
        )
        previous_checksum = try(
          terraform_data.script_versions.output[name].checksum,
          "NEW SCRIPT"
        )
        current_checksum = current.checksum
      }
      if current.checksum != try(terraform_data.script_versions.output[name].checksum, "")
    },
    {}
  )
}
```

### Pattern 3: Compliance and Audit Reporting

```hcl
# Define compliance requirements
variable "compliance_requirements" {
  type = object({
    required_categories = list(string)
    forbidden_shells    = list(string)
    max_timeout        = number
    required_platforms = list(string)
  })
  
  default = {
    required_categories = ["Security", "Compliance", "Monitoring"]
    forbidden_shells    = ["cmd"]  # Disallow batch scripts
    max_timeout        = 3600       # 1 hour maximum
    required_platforms = ["windows", "linux"]
  }
}

# Retrieve all scripts for compliance check
data "trmm_scripts" "compliance_audit" {}

# Perform compliance analysis
locals {
  compliance_analysis = {
    # Check required categories exist
    missing_categories = setsubtract(
      toset(var.compliance_requirements.required_categories),
      toset(distinct([for s in data.trmm_scripts.compliance_audit.scripts : s.category]))
    )
    
    # Find non-compliant scripts
    non_compliant_scripts = {
      forbidden_shell = [
        for s in data.trmm_scripts.compliance_audit.scripts : {
          name  = s.name
          shell = s.shell
        }
        if contains(var.compliance_requirements.forbidden_shells, s.shell)
      ]
      
      excessive_timeout = [
        for s in data.trmm_scripts.compliance_audit.scripts : {
          name    = s.name
          timeout = s.default_timeout
        }
        if s.default_timeout > var.compliance_requirements.max_timeout
      ]
      
      insufficient_platform_support = [
        for s in data.trmm_scripts.compliance_audit.scripts : {
          name      = s.name
          platforms = s.supported_platforms
        }
        if !alltrue([
          for p in var.compliance_requirements.required_platforms :
          contains(coalesce(s.supported_platforms, []), p)
        ])
      ]
    }
    
    # Summary statistics
    summary = {
      total_scripts          = length(data.trmm_scripts.compliance_audit.scripts)
      compliant_scripts      = length(data.trmm_scripts.compliance_audit.scripts) - 
                              length(local.compliance_analysis.non_compliant_scripts.forbidden_shell) -
                              length(local.compliance_analysis.non_compliant_scripts.excessive_timeout) -
                              length(local.compliance_analysis.non_compliant_scripts.insufficient_platform_support)
      compliance_percentage  = floor(
        (length(data.trmm_scripts.compliance_audit.scripts) - 
         length(local.compliance_analysis.non_compliant_scripts.forbidden_shell) -
         length(local.compliance_analysis.non_compliant_scripts.excessive_timeout) -
         length(local.compliance_analysis.non_compliant_scripts.insufficient_platform_support)
        ) * 100 / max(length(data.trmm_scripts.compliance_audit.scripts), 1)
      )
    }
  }
}

# Generate compliance report
output "compliance_report" {
  value = {
    analysis = local.compliance_analysis
    action_required = length(local.compliance_analysis.missing_categories) > 0 ||
                     length(local.compliance_analysis.non_compliant_scripts.forbidden_shell) > 0 ||
                     length(local.compliance_analysis.non_compliant_scripts.excessive_timeout) > 0 ||
                     length(local.compliance_analysis.non_compliant_scripts.insufficient_platform_support) > 0
  }
}
```

## Performance Optimization

### Query Optimization Strategies

1. **Filtered Queries**: Apply filters to reduce result set size
2. **Caching Pattern**: Store results in locals for multiple references
3. **Pagination Consideration**: Large result sets may require chunking

### Implementation Best Practices

```hcl
# Optimized query pattern
locals {
  # Cache filtered results
  maintenance_scripts_cache = data.trmm_scripts.maintenance_scripts.scripts
  
  # Pre-compute frequently used filters
  windows_maintenance = [
    for s in local.maintenance_scripts_cache : s
    if contains(coalesce(s.supported_platforms, []), "windows")
  ]
  
  linux_maintenance = [
    for s in local.maintenance_scripts_cache : s
    if contains(coalesce(s.supported_platforms, []), "linux")
  ]
}
```

## Error Handling

### Common Query Issues

1. **Empty Result Sets**
   - Verify filter criteria accuracy
   - Check for typos in category names
   - Ensure scripts exist matching criteria

2. **Performance Degradation**
   - Apply specific filters when possible
   - Avoid repeated full queries
   - Implement caching strategies

3. **Data Consistency**
   - Account for concurrent modifications
   - Implement retry logic for transient failures
   - Validate data completeness

## Related Resources

- [trmm_script](script.md) - Query individual scripts
- [Resource: trmm_script](../resources/script.md) - Manage scripts
- [trmm_script_snippets](script_snippets.md) - List script snippets

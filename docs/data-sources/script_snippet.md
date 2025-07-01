# trmm_script_snippet Data Source

## Overview

The `trmm_script_snippet` data source facilitates precise retrieval of individual code snippet resources within Tactical RMM, enabling modular script composition through systematic reference patterns.

## Technical Specifications

### Data Source Schema

```hcl
data "trmm_script_snippet" "example" {
  # Query Parameters (mutually exclusive)
  id   = number
  name = string
  
  # Computed Attributes
  desc  = string
  code  = string
  shell = string
}
```

### Query Parameters

| Parameter | Type | Description | Priority |
|-----------|------|-------------|----------|
| `id` | Number | Snippet identifier | Primary |
| `name` | String | Snippet name (exact match) | Secondary |

**Note**: Provide either `id` or `name`, not both. ID takes precedence if both supplied.

### Computed Attributes

All snippet attributes are exposed as computed values matching the resource schema.

## Implementation Patterns

### Pattern 1: Snippet Reference and Composition

```hcl
# Query snippets for script composition
data "trmm_script_snippet" "error_handler" {
  name = "ErrorHandler"
}

data "trmm_script_snippet" "logging_functions" {
  name = "LoggingFunctions"
}

data "trmm_script_snippet" "network_utils" {
  name = "NetworkUtils"
}

# Compose script using retrieved snippets
resource "trmm_script" "composite_script" {
  name        = "Comprehensive System Check"
  shell       = "powershell"
  category    = "Monitoring"
  
  script_body = <<-EOT
    # Include all utility snippets
    {{${data.trmm_script_snippet.error_handler.name}}}
    {{${data.trmm_script_snippet.logging_functions.name}}}
    {{${data.trmm_script_snippet.network_utils.name}}}
    
    # Main script logic using snippet functions
    Start-LogSession -ScriptName "System Check"
    
    try {
      # Network connectivity check
      if (-not (Test-NetworkConnectivity -Hostname "8.8.8.8" -Port 53)) {
        Write-LogEntry "DNS connectivity failed" -Level ERROR
        exit 1
      }
      
      Write-LogEntry "All checks passed" -Level INFO
      Stop-LogSession -ScriptName "System Check" -Success $true
    }
    catch {
      Write-LogEntry "Check failed: $_" -Level ERROR
      Stop-LogSession -ScriptName "System Check" -Success $false
      exit 2
    }
  EOT
  
  default_timeout = 120
}
```

### Pattern 2: Dynamic Snippet Selection

```hcl
# Environment-based snippet selection
variable "environment" {
  type    = string
  default = "production"
}

locals {
  snippet_mapping = {
    production = {
      config = "ProdConfigLoader"
      auth   = "ProdAuthentication"
      logger = "ProdLogging"
    }
    staging = {
      config = "StagingConfigLoader"
      auth   = "StagingAuthentication"
      logger = "StagingLogging"
    }
    development = {
      config = "DevConfigLoader"
      auth   = "DevAuthentication"
      logger = "DevLogging"
    }
  }
}

# Retrieve environment-specific snippets
data "trmm_script_snippet" "env_snippets" {
  for_each = local.snippet_mapping[var.environment]
  name     = each.value
}

# Validate all snippets exist
resource "null_resource" "snippet_validation" {
  for_each = data.trmm_script_snippet.env_snippets
  
  lifecycle {
    precondition {
      condition     = each.value.id != null
      error_message = "Required snippet '${each.key}' not found for ${var.environment} environment"
    }
  }
}
```

### Pattern 3: Snippet Inheritance and Extension

```hcl
# Base snippet retrieval
data "trmm_script_snippet" "base_functions" {
  name = "BaseFunctions"
}

# Create extended snippet with additional functionality
resource "trmm_script_snippet" "extended_functions" {
  name  = "ExtendedFunctions"
  desc  = "Extended functionality based on BaseFunctions"
  shell = data.trmm_script_snippet.base_functions.shell
  
  code = <<-EOT
    # Include base functions
    ${data.trmm_script_snippet.base_functions.code}
    
    # Extended functionality
    function Get-ExtendedSystemInfo {
      $BaseInfo = Get-SystemInfo  # From base snippet
      
      # Add extended properties
      $ExtendedInfo = $BaseInfo | Add-Member -NotePropertyMembers @{
        NetworkAdapters = Get-NetAdapter | Select-Object Name, Status
        InstalledUpdates = Get-HotFix | Select-Object -First 10
        RunningServices = Get-Service | Where-Object {$_.Status -eq 'Running'} | Measure-Object
      } -PassThru
      
      return $ExtendedInfo
    }
  EOT
}
```

### Pattern 4: Snippet Version Compatibility

```hcl
# Define snippet version requirements
locals {
  required_snippets = {
    logging = {
      name     = "LoggingV2"
      min_size = 500  # Minimum code size for validation
      required_functions = ["Write-LogEntry", "Start-LogSession", "Stop-LogSession"]
    }
    error_handling = {
      name     = "ErrorHandlerV2"
      min_size = 300
      required_functions = ["Set-ErrorHandler", "Get-LastError"]
    }
  }
}

# Retrieve and validate snippets
data "trmm_script_snippet" "versioned" {
  for_each = local.required_snippets
  name     = each.value.name
}

# Validate snippet compatibility
locals {
  snippet_validation = {
    for key, config in local.required_snippets : key => {
      found = data.trmm_script_snippet.versioned[key].id != null
      size_valid = length(data.trmm_script_snippet.versioned[key].code) >= config.min_size
      functions_present = alltrue([
        for func in config.required_functions :
        strcontains(data.trmm_script_snippet.versioned[key].code, func)
      ])
    }
  }
}

resource "null_resource" "compatibility_check" {
  for_each = local.snippet_validation
  
  lifecycle {
    precondition {
      condition = each.value.found && each.value.size_valid && each.value.functions_present
      error_message = "Snippet '${local.required_snippets[each.key].name}' does not meet compatibility requirements"
    }
  }
}
```

## Advanced Usage Patterns

### Pattern 1: Snippet Dependency Graph

```hcl
# Define snippet dependencies
locals {
  snippet_dependencies = {
    "AdvancedLogging" = ["BasicLogging", "FileOperations"]
    "NetworkDiagnostics" = ["NetworkUtils", "ErrorHandler"]
    "SecurityChecks" = ["RegistryUtils", "FilePermissions", "ServiceControl"]
  }
}

# Retrieve all snippets in dependency graph
data "trmm_script_snippet" "dependency_graph" {
  for_each = toset(flatten([
    for parent, deps in local.snippet_dependencies : concat([parent], deps)
  ]))
  
  name = each.value
}

# Build dependency resolution order
locals {
  resolution_order = flatten([
    for parent, deps in local.snippet_dependencies : concat(deps, [parent])
  ])
  
  unique_order = distinct(local.resolution_order)
}

# Generate combined snippet for complex operations
resource "trmm_script_snippet" "combined_utilities" {
  name  = "CombinedUtilities"
  desc  = "All utility functions combined"
  shell = "powershell"
  
  code = join("\n\n", [
    for snippet_name in local.unique_order :
    "# === ${snippet_name} ===\n${data.trmm_script_snippet.dependency_graph[snippet_name].code}"
  ])
}
```

### Pattern 2: Snippet Analysis and Documentation

```hcl
# Retrieve snippets for analysis
data "trmm_script_snippet" "analyzed" {
  for_each = toset(var.snippet_names)
  name     = each.value
}

# Analyze snippet characteristics
locals {
  snippet_analysis = {
    for name, snippet in data.trmm_script_snippet.analyzed : name => {
      shell = snippet.shell
      size  = length(snippet.code)
      lines = length(split("\n", snippet.code))
      
      # Language-specific analysis
      powershell_functions = snippet.shell == "powershell" ? 
        length(regexall("function\\s+[\\w-]+", snippet.code)) : 0
      
      python_functions = snippet.shell == "python" ? 
        length(regexall("def\\s+\\w+", snippet.code)) : 0
      
      # Complexity indicators
      has_error_handling = strcontains(lower(snippet.code), 
        snippet.shell == "powershell" ? "try" : "except"
      )
      
      uses_external_commands = snippet.shell == "powershell" ?
        strcontains(snippet.code, "Invoke-") : 
        strcontains(snippet.code, "subprocess")
    }
  }
}

# Generate documentation
output "snippet_documentation" {
  value = {
    for name, analysis in local.snippet_analysis : name => {
      description = data.trmm_script_snippet.analyzed[name].desc
      metrics = analysis
      usage_hint = format(
        "Use {{%s}} in %s scripts for %s",
        name,
        analysis.shell,
        data.trmm_script_snippet.analyzed[name].desc
      )
    }
  }
}
```

### Pattern 3: Cross-Reference Validation

```hcl
# Scripts that should use specific snippets
variable "script_snippet_requirements" {
  type = map(list(string))
  default = {
    "Database Maintenance" = ["DatabaseUtils", "ErrorHandler"]
    "Security Audit"       = ["SecurityChecks", "ReportGenerator"]
    "System Monitoring"    = ["SystemMetrics", "AlertingFunctions"]
  }
}

# Retrieve required snippets
data "trmm_script_snippet" "required" {
  for_each = toset(flatten(values(var.script_snippet_requirements)))
  name     = each.value
}

# Retrieve scripts to validate
data "trmm_script" "validation_targets" {
  for_each = var.script_snippet_requirements
  name     = each.key
}

# Validate snippet usage
locals {
  snippet_usage_validation = {
    for script_name, required_snippets in var.script_snippet_requirements : script_name => {
      script_found = can(data.trmm_script.validation_targets[script_name].id)
      
      missing_snippets = script_found ? [
        for snippet in required_snippets :
        snippet if !strcontains(
          data.trmm_script.validation_targets[script_name].script_body,
          "{{${snippet}}}"
        )
      ] : []
      
      all_snippets_exist = alltrue([
        for snippet in required_snippets :
        can(data.trmm_script_snippet.required[snippet].id)
      ])
    }
  }
}

output "snippet_compliance_report" {
  value = {
    for script, validation in local.snippet_usage_validation :
    script => {
      compliant = validation.script_found && 
                 length(validation.missing_snippets) == 0 && 
                 validation.all_snippets_exist
      issues = concat(
        validation.script_found ? [] : ["Script not found"],
        validation.missing_snippets,
        validation.all_snippets_exist ? [] : ["Some required snippets don't exist"]
      )
    }
  }
}
```

## Error Handling

### Common Query Errors

1. **Snippet Not Found**
   ```
   Error: Snippet with name "NonExistent" not found
   ```
   - Verify exact snippet name match
   - Check for case sensitivity
   - Ensure snippet exists in environment

2. **Access Denied**
   ```
   Error: Insufficient permissions to read snippet
   ```
   - Verify API key permissions
   - Check snippet access policies
   - Review organizational restrictions

3. **Malformed Query**
   ```
   Error: Invalid query parameters
   ```
   - Provide either id or name, not both
   - Ensure parameter types are correct
   - Validate query syntax

### Diagnostic Techniques

```hcl
# Diagnostic snippet queries
resource "terraform_data" "snippet_diagnostics" {
  input = {
    timestamp = timestamp()
    snippets_checked = keys(data.trmm_script_snippet.analyzed)
    snippets_found = [
      for name, snippet in data.trmm_script_snippet.analyzed :
      name if snippet.id != null
    ]
    query_errors = [
      for name, snippet in data.trmm_script_snippet.analyzed :
      name if snippet.id == null
    ]
  }
}

output "diagnostic_results" {
  value = terraform_data.snippet_diagnostics.output
}
```

## Best Practices

### 1. Query Efficiency
- Cache snippet lookups in locals
- Minimize redundant queries
- Use ID queries for performance

### 2. Validation Patterns
- Verify snippet existence before use
- Validate snippet compatibility
- Check for required functions

### 3. Documentation Standards
- Document snippet dependencies
- Maintain snippet versioning
- Include usage examples

## Related Resources

- [trmm_script_snippets](script_snippets.md) - List multiple snippets
- [Resource: trmm_script_snippet](../resources/script_snippet.md) - Manage snippets
- [trmm_script](script.md) - Scripts using snippets

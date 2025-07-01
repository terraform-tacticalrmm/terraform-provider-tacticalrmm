# tacticalrmm_script Resource

## Overview

The `tacticalrmm_script` resource manages automation scripts within Tactical RMM, supporting PowerShell, Python, Shell, and Batch script deployment across Windows, Linux, and macOS endpoints.

## Technical Specifications

### Resource Schema

```hcl
resource "tacticalrmm_script" "example" {
  # Required Attributes
  name        = string
  shell       = string
  script_body = string
  
  # Optional Attributes
  description          = string
  category            = string
  default_timeout     = number
  favorite            = bool
  hidden              = bool
  run_as_user         = bool
  syntax              = string
  args                = list(string)
  env_vars            = list(string)
  supported_platforms = list(string)
  
  # Computed Attributes
  id          = number
  script_type = string
}
```

### Attribute Reference

#### Required Attributes

| Attribute | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| `name` | String | Script identifier name | Unique, max 255 characters |
| `shell` | String | Execution environment | `powershell`, `cmd`, `python`, `shell`, `nushell`, `deno` |
| `script_body` | String | Script content | Valid syntax for specified shell |

#### Optional Attributes

| Attribute | Type | Description | Default | Constraints |
|-----------|------|-------------|---------|-------------|
| `description` | String | Script purpose description | `null` | Max 200 characters |
| `category` | String | Organizational category | `null` | Custom categorization |
| `default_timeout` | Number | Execution timeout (seconds) | `90` | Range: 1-86400 |
| `favorite` | Bool | Favorite status flag | `false` | - |
| `hidden` | Bool | Hidden from UI lists | `false` | - |
| `run_as_user` | Bool | Execute as logged-in user | `false` | Windows only |
| `syntax` | String | Syntax highlighting hint | `null` | Editor optimization |
| `args` | List(String) | Command-line arguments | `null` | Shell-specific formatting |
| `env_vars` | List(String) | Environment variables | `null` | `KEY=VALUE` format |
| `supported_platforms` | List(String) | Target platforms | `null` | `windows`, `linux`, `darwin` |

#### Computed Attributes

| Attribute | Type | Description | Value |
|-----------|------|-------------|-------|
| `id` | Number | Resource identifier | Auto-generated |
| `script_type` | String | Script classification | `userdefined` |

## Implementation Examples

### Example 1: Windows PowerShell Maintenance Script

```hcl
resource "tacticalrmm_script" "windows_maintenance" {
  name        = "Windows System Maintenance"
  description = "Comprehensive system maintenance tasks"
  shell       = "powershell"
  category    = "Maintenance"
  
  script_body = <<-EOT
    # System Maintenance Script
    param(
      [int]$DaysToKeep = 30,
      [switch]$Verbose
    )
    
    # Clear temporary files
    $TempPaths = @(
      "$env:TEMP",
      "$env:WINDIR\Temp",
      "$env:WINDIR\Prefetch"
    )
    
    foreach ($Path in $TempPaths) {
      if (Test-Path $Path) {
        Get-ChildItem -Path $Path -Recurse -Force -ErrorAction SilentlyContinue |
          Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-$DaysToKeep) } |
          Remove-Item -Force -Recurse -ErrorAction SilentlyContinue
      }
    }
    
    # Clear event logs
    wevtutil enum-logs | ForEach-Object {
      wevtutil clear-log $_ /quiet 2>$null
    }
    
    if ($Verbose) {
      Write-Output "Maintenance completed at $(Get-Date)"
    }
  EOT
  
  default_timeout     = 600
  run_as_user        = false
  supported_platforms = ["windows"]
  
  args = [
    "-ExecutionPolicy", "Bypass",
    "-NoProfile"
  ]
  
  env_vars = [
    "MAINTENANCE_MODE=true",
    "LOG_LEVEL=INFO"
  ]
}
```

### Example 2: Cross-Platform Python Monitoring Script

```hcl
resource "tacticalrmm_script" "system_monitor" {
  name        = "System Resource Monitor"
  description = "Monitor CPU, memory, and disk usage"
  shell       = "python"
  category    = "Monitoring"
  
  script_body = file("${path.module}/scripts/system_monitor.py")
  
  default_timeout     = 60
  supported_platforms = ["windows", "linux", "darwin"]
  
  env_vars = [
    "ALERT_THRESHOLD_CPU=90",
    "ALERT_THRESHOLD_MEM=85",
    "ALERT_THRESHOLD_DISK=90"
  ]
}
```

### Example 3: Linux Shell Configuration Script

```hcl
resource "tacticalrmm_script" "linux_security_hardening" {
  name        = "Linux Security Hardening"
  description = "Apply security hardening configurations"
  shell       = "shell"
  category    = "Security"
  
  script_body = <<-EOT
    #!/bin/bash
    set -euo pipefail
    
    # Disable unused network protocols
    echo "install dccp /bin/true" >> /etc/modprobe.d/disable-protocols.conf
    echo "install sctp /bin/true" >> /etc/modprobe.d/disable-protocols.conf
    echo "install rds /bin/true" >> /etc/modprobe.d/disable-protocols.conf
    echo "install tipc /bin/true" >> /etc/modprobe.d/disable-protocols.conf
    
    # Configure kernel parameters
    cat >> /etc/sysctl.d/99-security.conf <<EOF
    net.ipv4.ip_forward = 0
    net.ipv4.conf.all.accept_source_route = 0
    net.ipv4.conf.all.accept_redirects = 0
    net.ipv4.conf.all.secure_redirects = 0
    net.ipv4.conf.all.log_martians = 1
    net.ipv4.conf.default.log_martians = 1
    net.ipv4.icmp_echo_ignore_broadcasts = 1
    net.ipv4.icmp_ignore_bogus_error_responses = 1
    net.ipv4.tcp_syncookies = 1
    net.ipv4.conf.all.send_redirects = 0
    net.ipv4.conf.default.send_redirects = 0
    EOF
    
    # Apply settings
    sysctl -p /etc/sysctl.d/99-security.conf
    
    echo "Security hardening completed"
  EOT
  
  default_timeout     = 300
  run_as_user        = false
  supported_platforms = ["linux"]
}
```

### Example 4: Script with Dynamic Content

```hcl
# Using templatefile for dynamic script generation
resource "tacticalrmm_script" "dynamic_config" {
  name        = "Dynamic Configuration Script"
  description = "Configures system based on variables"
  shell       = "powershell"
  category    = "Configuration"
  
  script_body = templatefile("${path.module}/templates/config.ps1.tpl", {
    domain_controller = var.domain_controller
    dns_servers       = var.dns_servers
    ntp_servers       = var.ntp_servers
  })
  
  default_timeout = 180
}
```

## State Management

### Import Existing Scripts

```bash
# Import by script ID
terraform import tacticalrmm_script.example 123
```

### State Attributes

The provider maintains precise state management:
- **Null Preservation**: Distinguishes between null and empty arrays
- **Computed Defaults**: Automatically populates server-side defaults
- **Change Detection**: Tracks all attribute modifications

## Implementation Patterns

### Script Organization Strategy

```hcl
# Organize scripts by category using locals
locals {
  script_categories = {
    maintenance = "Maintenance"
    monitoring  = "Monitoring"
    security    = "Security"
    deployment  = "Deployment"
  }
}

resource "tacticalrmm_script" "organized_scripts" {
  for_each = var.scripts
  
  name        = each.value.name
  description = each.value.description
  shell       = each.value.shell
  category    = local.script_categories[each.value.category]
  script_body = file("${path.module}/scripts/${each.key}.${each.value.extension}")
  
  default_timeout     = each.value.timeout
  supported_platforms = each.value.platforms
}
```

### Error Handling Patterns

```hcl
resource "tacticalrmm_script" "error_handling_example" {
  name        = "Robust Error Handler"
  shell       = "powershell"
  category    = "Utilities"
  
  script_body = <<-EOT
    try {
      # Main script logic
      $Result = Invoke-WebRequest -Uri $env:CHECK_URL -TimeoutSec 30
      
      if ($Result.StatusCode -eq 200) {
        Write-Output "Success: Service is responsive"
        exit 0
      } else {
        Write-Error "Service returned status: $($Result.StatusCode)"
        exit 1
      }
    }
    catch {
      Write-Error "Script execution failed: $_"
      exit 2
    }
    finally {
      # Cleanup operations
      Remove-Variable -Name Result -ErrorAction SilentlyContinue
    }
  EOT
  
  default_timeout = 60
  
  env_vars = [
    "CHECK_URL=https://status.example.com/api/health"
  ]
}
```

## Best Practices

### 1. Script Modularity
- Use script snippets for reusable code components
- Reference snippets using `{{SnippetName}}` syntax
- Maintain single-responsibility principle

### 2. Platform Targeting
```hcl
# Explicitly define supported platforms
supported_platforms = ["windows", "linux"]  # Prevents execution on unsupported systems
```

### 3. Timeout Configuration
- Set realistic timeouts based on script complexity
- Consider network latency for remote operations
- Add buffer for variable system performance

### 4. Security Considerations
- Avoid hardcoding sensitive data
- Use `tacticalrmm_keystore` for credential storage
- Implement proper error handling to prevent information disclosure

## Troubleshooting

### Common Issues

1. **Script Creation Failures**
   - Verify unique script names
   - Validate shell type compatibility
   - Check script syntax validity

2. **Execution Timeouts**
   - Increase `default_timeout` for long-running operations
   - Optimize script performance
   - Consider breaking into smaller scripts

3. **Platform Compatibility**
   - Ensure shell type matches target platform
   - Test scripts on representative systems
   - Use conditional logic for cross-platform scripts

### Debug Techniques

```hcl
# Enable verbose output for debugging
resource "tacticalrmm_script" "debug_example" {
  name = "Debug Script"
  shell = "powershell"
  
  script_body = <<-EOT
    $VerbosePreference = 'Continue'
    $DebugPreference = 'Continue'
    
    Write-Verbose "Starting script execution"
    Write-Debug "Variable state: $PSBoundParameters"
    
    # Script logic here
  EOT
  
  default_timeout = 120
}
```

## Related Resources

- [tacticalrmm_script_snippet](script_snippet.md) - Reusable code components
- [tacticalrmm_keystore](keystore.md) - Secure credential storage
- [Data Source: tacticalrmm_script](../data-sources/script.md) - Query existing scripts

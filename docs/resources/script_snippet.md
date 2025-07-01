# tacticalrmm_script_snippet Resource

## Overview

The `tacticalrmm_script_snippet` resource manages reusable code components within Tactical RMM, enabling modular script composition through template substitution in parent scripts.

## Technical Specifications

### Resource Schema

```hcl
resource "tacticalrmm_script_snippet" "example" {
  # Required Attributes
  name = string
  code = string
  
  # Optional Attributes
  desc  = string
  shell = string
  
  # Computed Attributes
  id = number
}
```

### Attribute Reference

#### Required Attributes

| Attribute | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| `name` | String | Snippet identifier | Unique, max 40 characters, no spaces |
| `code` | String | Snippet content | Valid code for target shell |

#### Optional Attributes

| Attribute | Type | Description | Default | Constraints |
|-----------|------|-------------|---------|-------------|
| `desc` | String | Snippet description | `null` | Max 50 characters |
| `shell` | String | Target shell type | `powershell` | `powershell`, `cmd`, `python`, `shell` |

#### Computed Attributes

| Attribute | Type | Description | Value |
|-----------|------|-------------|-------|
| `id` | Number | Resource identifier | Auto-generated |

## Implementation Architecture

### Snippet Integration Pattern

Scripts reference snippets using the `{{SnippetName}}` syntax:

```hcl
# Define reusable snippet
resource "tacticalrmm_script_snippet" "error_handler" {
  name  = "ErrorHandler"
  desc  = "Standard error handling logic"
  shell = "powershell"
  
  code = <<-EOT
    trap {
      Write-Error "An error occurred: $_"
      Write-Error $_.ScriptStackTrace
      exit 1
    }
    $ErrorActionPreference = 'Stop'
  EOT
}

# Use snippet in script
resource "tacticalrmm_script" "maintenance_task" {
  name        = "System Maintenance"
  shell       = "powershell"
  
  script_body = <<-EOT
    # Include error handling
    {{ErrorHandler}}
    
    # Main script logic
    Write-Output "Starting maintenance tasks..."
    # ... rest of script
  EOT
  
  depends_on = [tacticalrmm_script_snippet.error_handler]
}
```

## Implementation Examples

### Example 1: PowerShell Function Library

```hcl
resource "tacticalrmm_script_snippet" "powershell_logging" {
  name  = "LoggingFunctions"
  desc  = "Standardized logging functions"
  shell = "powershell"
  
  code = <<-EOT
    function Write-LogEntry {
      param(
        [Parameter(Mandatory=$true)]
        [string]$Message,
        
        [ValidateSet('INFO','WARN','ERROR','DEBUG')]
        [string]$Level = 'INFO',
        
        [string]$LogPath = "$env:TEMP\trmm-script.log"
      )
      
      $Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
      $LogEntry = "[$Timestamp] [$Level] $Message"
      
      # Console output with color coding
      switch ($Level) {
        'ERROR' { Write-Host $LogEntry -ForegroundColor Red }
        'WARN'  { Write-Host $LogEntry -ForegroundColor Yellow }
        'DEBUG' { Write-Host $LogEntry -ForegroundColor Gray }
        default { Write-Host $LogEntry }
      }
      
      # File output
      Add-Content -Path $LogPath -Value $LogEntry -Force
    }
    
    function Start-LogSession {
      param([string]$ScriptName)
      Write-LogEntry -Message "Starting script: $ScriptName" -Level INFO
      Write-LogEntry -Message "User: $env:USERNAME, Computer: $env:COMPUTERNAME" -Level DEBUG
    }
    
    function Stop-LogSession {
      param(
        [string]$ScriptName,
        [bool]$Success = $true
      )
      $Status = if ($Success) { "SUCCESS" } else { "FAILED" }
      Write-LogEntry -Message "Script $ScriptName completed with status: $Status" -Level INFO
    }
  EOT
}
```

### Example 2: Python Utility Functions

```hcl
resource "tacticalrmm_script_snippet" "python_utilities" {
  name  = "PythonUtils"
  desc  = "Common Python utility functions"
  shell = "python"
  
  code = <<-EOT
    import os
    import sys
    import json
    import logging
    from datetime import datetime
    from typing import Dict, List, Any, Optional
    
    # Configure logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    logger = logging.getLogger(__name__)
    
    def safe_json_load(file_path: str) -> Optional[Dict[str, Any]]:
        """Safely load JSON file with error handling."""
        try:
            with open(file_path, 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            logger.error(f"File not found: {file_path}")
        except json.JSONDecodeError as e:
            logger.error(f"Invalid JSON in {file_path}: {e}")
        except Exception as e:
            logger.error(f"Unexpected error reading {file_path}: {e}")
        return None
    
    def get_system_info() -> Dict[str, Any]:
        """Gather basic system information."""
        import platform
        import socket
        
        return {
            'hostname': socket.gethostname(),
            'platform': platform.system(),
            'platform_release': platform.release(),
            'platform_version': platform.version(),
            'architecture': platform.machine(),
            'processor': platform.processor(),
            'python_version': platform.python_version(),
            'timestamp': datetime.utcnow().isoformat()
        }
    
    def ensure_directory(path: str) -> bool:
        """Ensure directory exists, create if necessary."""
        try:
            os.makedirs(path, exist_ok=True)
            return True
        except Exception as e:
            logger.error(f"Failed to create directory {path}: {e}")
            return False
  EOT
}
```

### Example 3: Shell Script Common Functions

```hcl
resource "tacticalrmm_script_snippet" "shell_common" {
  name  = "ShellCommon"
  desc  = "Common shell script functions"
  shell = "shell"
  
  code = <<-EOT
    # Color codes for output
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m' # No Color
    
    # Logging functions
    log_info() {
      echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    }
    
    log_warn() {
      echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    }
    
    log_error() {
      echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" >&2
    }
    
    # Check if running as root
    require_root() {
      if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root"
        exit 1
      fi
    }
    
    # Check if command exists
    command_exists() {
      command -v "$1" >/dev/null 2>&1
    }
    
    # Safe command execution with logging
    safe_execute() {
      local cmd="$1"
      log_info "Executing: $cmd"
      
      if output=$(eval "$cmd" 2>&1); then
        log_info "Command succeeded"
        echo "$output"
        return 0
      else
        log_error "Command failed with exit code $?"
        echo "$output" >&2
        return 1
      fi
    }
    
    # Backup file before modification
    backup_file() {
      local file="$1"
      local backup="${file}.backup.$(date +%Y%m%d_%H%M%S)"
      
      if [ -f "$file" ]; then
        cp -p "$file" "$backup"
        log_info "Created backup: $backup"
      else
        log_warn "File not found for backup: $file"
      fi
    }
  EOT
}
```

### Example 4: Multi-Snippet Composition

```hcl
# Define multiple complementary snippets
resource "tacticalrmm_script_snippet" "config_validator" {
  name  = "ConfigValidator"
  desc  = "Configuration validation logic"
  shell = "powershell"
  
  code = <<-EOT
    function Test-Configuration {
      param($Config)
      
      $RequiredKeys = @('ServerName', 'Port', 'Protocol')
      $MissingKeys = $RequiredKeys | Where-Object { -not $Config.ContainsKey($_) }
      
      if ($MissingKeys.Count -gt 0) {
        throw "Missing required configuration keys: $($MissingKeys -join ', ')"
      }
      
      if ($Config.Port -lt 1 -or $Config.Port -gt 65535) {
        throw "Invalid port number: $($Config.Port)"
      }
      
      return $true
    }
  EOT
}

resource "tacticalrmm_script_snippet" "network_utilities" {
  name  = "NetworkUtils"
  desc  = "Network testing utilities"
  shell = "powershell"
  
  code = <<-EOT
    function Test-NetworkConnectivity {
      param(
        [string]$Hostname,
        [int]$Port,
        [int]$TimeoutSeconds = 5
      )
      
      try {
        $TCPClient = New-Object System.Net.Sockets.TcpClient
        $Connect = $TCPClient.BeginConnect($Hostname, $Port, $null, $null)
        $Wait = $Connect.AsyncWaitHandle.WaitOne($TimeoutSeconds * 1000, $false)
        
        if ($Wait) {
          $TCPClient.EndConnect($Connect)
          $TCPClient.Close()
          return $true
        } else {
          $TCPClient.Close()
          return $false
        }
      } catch {
        return $false
      }
    }
  EOT
}

# Composite script using multiple snippets
resource "tacticalrmm_script" "service_health_check" {
  name        = "Service Health Check"
  shell       = "powershell"
  category    = "Monitoring"
  
  script_body = <<-EOT
    # Include all utilities
    {{LoggingFunctions}}
    {{ConfigValidator}}
    {{NetworkUtils}}
    
    # Main health check logic
    Start-LogSession -ScriptName "Service Health Check"
    
    try {
      # Load configuration
      $ConfigPath = "$env:ProgramData\TRMM\service-config.json"
      $Config = Get-Content $ConfigPath | ConvertFrom-Json -AsHashtable
      
      # Validate configuration
      Write-LogEntry "Validating configuration..." -Level INFO
      Test-Configuration -Config $Config
      
      # Test connectivity
      Write-LogEntry "Testing network connectivity..." -Level INFO
      $Connected = Test-NetworkConnectivity `
        -Hostname $Config.ServerName `
        -Port $Config.Port
      
      if ($Connected) {
        Write-LogEntry "Service is reachable" -Level INFO
        Stop-LogSession -ScriptName "Service Health Check" -Success $true
        exit 0
      } else {
        Write-LogEntry "Service is not reachable" -Level ERROR
        Stop-LogSession -ScriptName "Service Health Check" -Success $false
        exit 1
      }
    } catch {
      Write-LogEntry "Health check failed: $_" -Level ERROR
      Stop-LogSession -ScriptName "Service Health Check" -Success $false
      exit 2
    }
  EOT
  
  default_timeout = 30
  
  depends_on = [
    tacticalrmm_script_snippet.powershell_logging,
    tacticalrmm_script_snippet.config_validator,
    tacticalrmm_script_snippet.network_utilities
  ]
}
```

## State Management

### Import Existing Snippets

```bash
# Import by snippet ID
terraform import tacticalrmm_script_snippet.example 456
```

### State Synchronization

The provider maintains accurate state through:
- Unique name validation
- Change detection for all attributes
- Proper null value handling

## Best Practices

### 1. Naming Conventions

```hcl
# Use consistent, descriptive naming
locals {
  snippet_prefix = "Org"  # Organization prefix
  
  snippets = {
    logging   = "${local.snippet_prefix}Logging"
    auth      = "${local.snippet_prefix}Authentication"
    network   = "${local.snippet_prefix}Network"
    storage   = "${local.snippet_prefix}Storage"
  }
}
```

### 2. Version Control Integration

```hcl
# Load snippets from version-controlled files
resource "tacticalrmm_script_snippet" "from_file" {
  for_each = fileset("${path.module}/snippets", "*.ps1")
  
  name  = replace(each.key, ".ps1", "")
  desc  = "Loaded from ${each.key}"
  shell = "powershell"
  code  = file("${path.module}/snippets/${each.key}")
}
```

### 3. Documentation Standards

```hcl
resource "tacticalrmm_script_snippet" "documented_example" {
  name  = "DocumentedFunction"
  desc  = "Example with inline documentation"
  shell = "powershell"
  
  code = <<-EOT
    <#
    .SYNOPSIS
        Performs automated system validation
    
    .DESCRIPTION
        This function validates system configuration against
        predefined standards and reports compliance status
    
    .PARAMETER ConfigPath
        Path to configuration file
    
    .EXAMPLE
        Test-SystemCompliance -ConfigPath "C:\config\standards.json"
    #>
    function Test-SystemCompliance {
      # Implementation here
    }
  EOT
}
```

### 4. Error Handling Patterns

```hcl
# Snippet with comprehensive error handling
resource "tacticalrmm_script_snippet" "safe_operations" {
  name  = "SafeOperations"
  shell = "python"
  
  code = <<-EOT
    import functools
    import traceback
    
    def safe_operation(func):
        """Decorator for safe function execution with logging."""
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            try:
                logger.info(f"Executing {func.__name__}")
                result = func(*args, **kwargs)
                logger.info(f"{func.__name__} completed successfully")
                return result
            except Exception as e:
                logger.error(f"{func.__name__} failed: {str(e)}")
                logger.debug(traceback.format_exc())
                raise
        return wrapper
  EOT
}
```

## Troubleshooting

### Common Issues

1. **Name Conflicts**
   - Ensure unique snippet names across environment
   - Use namespace prefixes for organization
   - Check for case sensitivity

2. **Template Substitution Failures**
   - Verify exact snippet name match
   - Ensure snippet is created before script
   - Use `depends_on` for explicit ordering

3. **Shell Compatibility**
   - Match snippet shell type with parent script
   - Consider cross-shell compatibility needs
   - Test snippets independently

### Validation Techniques

```hcl
# Validate snippet functionality
resource "tacticalrmm_script" "snippet_test" {
  name = "Test_${tacticalrmm_script_snippet.example.name}"
  shell = tacticalrmm_script_snippet.example.shell
  
  script_body = <<-EOT
    # Test snippet inclusion
    {{${tacticalrmm_script_snippet.example.name}}}
    
    # Verify snippet functionality
    # Add test cases here
  EOT
  
  category = "Testing"
  hidden   = true
}
```

## Related Resources

- [tacticalrmm_script](script.md) - Parent script resources
- [Data Source: tacticalrmm_script_snippet](../data-sources/script_snippet.md) - Query existing snippets
- [Data Source: tacticalrmm_script_snippets](../data-sources/script_snippets.md) - List all snippets

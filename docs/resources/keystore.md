# tacticalrmm_keystore Resource

## Overview

The `tacticalrmm_keystore` resource provides secure key-value storage within Tactical RMM, enabling centralized management of configuration parameters, credentials, and sensitive data with controlled access patterns.

## Technical Specifications

### Resource Schema

```hcl
resource "tacticalrmm_keystore" "example" {
  # Required Attributes
  name  = string
  value = string
  
  # Computed Attributes
  id = number
}
```

### Attribute Reference

#### Required Attributes

| Attribute | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| `name` | String | Unique key identifier | Max 25 characters, alphanumeric with underscores |
| `value` | String | Stored value (sensitive) | Encrypted at rest, no size limit |

#### Computed Attributes

| Attribute | Type | Description | Value |
|-----------|------|-------------|-------|
| `id` | Number | Resource identifier | Auto-generated, immutable |

## Implementation Architecture

### Security Model

The keystore implements a secure storage pattern:

1. **Encryption**: Values encrypted at rest in database
2. **Access Control**: API key permissions enforced
3. **Audit Trail**: Access logging for compliance
4. **Terraform Integration**: Sensitive value marking prevents exposure

### Value Retrieval Pattern

Scripts access keystore values through the Tactical RMM agent:

```powershell
# PowerShell example
$apiKey = Get-TrmmKeystore -Name "external_api_key"

# Python example
api_key = trmm.get_keystore("external_api_key")

# Shell example
API_KEY=$(trmm-keystore get external_api_key)
```

## Implementation Examples

### Example 1: API Credentials Management

```hcl
# Store multiple API credentials systematically
locals {
  api_credentials = {
    smtp_server = {
      username = var.smtp_username
      password = var.smtp_password
    }
    monitoring_api = {
      endpoint = "https://api.monitoring.service.com/v2"
      key      = var.monitoring_api_key
      secret   = var.monitoring_api_secret
    }
    backup_service = {
      access_key = var.backup_access_key
      secret_key = var.backup_secret_key
      region     = "us-east-1"
    }
  }
}

# SMTP Credentials
resource "tacticalrmm_keystore" "smtp_username" {
  name  = "smtp_username"
  value = local.api_credentials.smtp_server.username
}

resource "tacticalrmm_keystore" "smtp_password" {
  name  = "smtp_password"
  value = local.api_credentials.smtp_server.password
}

# Monitoring API Configuration
resource "tacticalrmm_keystore" "monitoring_endpoint" {
  name  = "monitoring_api_endpoint"
  value = local.api_credentials.monitoring_api.endpoint
}

resource "tacticalrmm_keystore" "monitoring_key" {
  name  = "monitoring_api_key"
  value = local.api_credentials.monitoring_api.key
}

resource "tacticalrmm_keystore" "monitoring_secret" {
  name  = "monitoring_api_secret"
  value = local.api_credentials.monitoring_api.secret
}

# Backup Service Credentials
resource "tacticalrmm_keystore" "backup_credentials" {
  for_each = local.api_credentials.backup_service
  
  name  = "backup_${each.key}"
  value = each.value
}
```

### Example 2: Environment-Specific Configuration

```hcl
# Dynamic configuration based on workspace
locals {
  environment = terraform.workspace
  
  env_configs = {
    production = {
      db_connection_string = "Server=prod-db.internal;Database=TacticalRMM;Integrated Security=true;"
      log_level            = "WARNING"
      feature_flags        = jsonencode({
        enable_beta_features = false
        maintenance_mode     = false
        debug_logging       = false
      })
    }
    staging = {
      db_connection_string = "Server=staging-db.internal;Database=TacticalRMM_Staging;Integrated Security=true;"
      log_level            = "INFO"
      feature_flags        = jsonencode({
        enable_beta_features = true
        maintenance_mode     = false
        debug_logging       = true
      })
    }
    development = {
      db_connection_string = "Server=localhost;Database=TacticalRMM_Dev;User Id=dev;Password=dev123;"
      log_level            = "DEBUG"
      feature_flags        = jsonencode({
        enable_beta_features = true
        maintenance_mode     = true
        debug_logging       = true
      })
    }
  }
}

# Store environment-specific configurations
resource "tacticalrmm_keystore" "env_config" {
  for_each = local.env_configs[local.environment]
  
  name  = "${local.environment}_${each.key}"
  value = each.value
}

# Reference in scripts
resource "tacticalrmm_script" "config_reader" {
  name        = "Read Environment Configuration"
  shell       = "powershell"
  category    = "Configuration"
  
  script_body = <<-EOT
    $Environment = "${local.environment}"
    
    # Retrieve configuration values
    $ConnectionString = Get-TrmmKeystore -Name "${local.environment}_db_connection_string"
    $LogLevel = Get-TrmmKeystore -Name "${local.environment}_log_level"
    $FeatureFlags = Get-TrmmKeystore -Name "${local.environment}_feature_flags" | ConvertFrom-Json
    
    Write-Output "Environment: $Environment"
    Write-Output "Log Level: $LogLevel"
    Write-Output "Beta Features: $($FeatureFlags.enable_beta_features)"
  EOT
  
  depends_on = [tacticalrmm_keystore.env_config]
}
```

### Example 3: Certificate and Key Management

```hcl
# Store certificates and private keys
resource "tacticalrmm_keystore" "ssl_certificate" {
  name  = "wildcard_cert_pem"
  value = file("${path.module}/certs/wildcard.crt")
}

resource "tacticalrmm_keystore" "ssl_private_key" {
  name  = "wildcard_key_pem"
  value = file("${path.module}/certs/wildcard.key")
}

resource "tacticalrmm_keystore" "ca_bundle" {
  name  = "ca_bundle_pem"
  value = file("${path.module}/certs/ca-bundle.crt")
}

# Script to deploy certificates
resource "tacticalrmm_script" "deploy_certificates" {
  name        = "Deploy SSL Certificates"
  shell       = "powershell"
  category    = "Security"
  
  script_body = <<-EOT
    # Retrieve certificates from keystore
    $CertPEM = Get-TrmmKeystore -Name "wildcard_cert_pem"
    $KeyPEM = Get-TrmmKeystore -Name "wildcard_key_pem"
    $CABundle = Get-TrmmKeystore -Name "ca_bundle_pem"
    
    # Define certificate paths
    $CertPath = "C:\Certificates\wildcard.crt"
    $KeyPath = "C:\Certificates\wildcard.key"
    $CAPath = "C:\Certificates\ca-bundle.crt"
    
    # Ensure directory exists
    New-Item -ItemType Directory -Path "C:\Certificates" -Force | Out-Null
    
    # Deploy certificates with proper permissions
    Set-Content -Path $CertPath -Value $CertPEM -Encoding UTF8
    Set-Content -Path $KeyPath -Value $KeyPEM -Encoding UTF8
    Set-Content -Path $CAPath -Value $CABundle -Encoding UTF8
    
    # Secure private key
    $Acl = Get-Acl $KeyPath
    $Acl.SetAccessRuleProtection($true, $false)
    $AdminRule = New-Object System.Security.AccessControl.FileSystemAccessRule(
      "BUILTIN\Administrators", "FullControl", "Allow"
    )
    $SystemRule = New-Object System.Security.AccessControl.FileSystemAccessRule(
      "NT AUTHORITY\SYSTEM", "FullControl", "Allow"
    )
    $Acl.SetAccessRule($AdminRule)
    $Acl.SetAccessRule($SystemRule)
    Set-Acl -Path $KeyPath -AclObject $Acl
    
    Write-Output "Certificates deployed successfully"
  EOT
  
  default_timeout = 60
  run_as_user    = false
  
  depends_on = [
    tacticalrmm_keystore.ssl_certificate,
    tacticalrmm_keystore.ssl_private_key,
    tacticalrmm_keystore.ca_bundle
  ]
}
```

### Example 4: Dynamic Secret Rotation

```hcl
# Implement secret rotation pattern
variable "rotation_timestamp" {
  description = "Timestamp for secret rotation tracking"
  type        = string
  default     = ""
}

locals {
  # Generate rotation suffix based on timestamp
  rotation_suffix = var.rotation_timestamp != "" ? "_${var.rotation_timestamp}" : ""
  
  # Define secrets requiring rotation
  rotating_secrets = {
    service_account_password = random_password.service_account.result
    api_signing_key         = random_password.api_signing.result
    encryption_key          = random_password.encryption.result
  }
}

# Generate random passwords
resource "random_password" "service_account" {
  length  = 32
  special = true
}

resource "random_password" "api_signing" {
  length  = 64
  special = false  # Base64 friendly
}

resource "random_password" "encryption" {
  length  = 32
  special = false
}

# Store with rotation support
resource "tacticalrmm_keystore" "rotating_secrets" {
  for_each = local.rotating_secrets
  
  name  = "${each.key}${local.rotation_suffix}"
  value = each.value
}

# Cleanup script for old secrets
resource "tacticalrmm_script" "cleanup_old_secrets" {
  name        = "Cleanup Rotated Secrets"
  shell       = "python"
  category    = "Maintenance"
  
  script_body = <<-EOT
    import re
    from datetime import datetime, timedelta
    
    # Get all keystore entries
    keystores = trmm.list_keystores()
    
    # Pattern for rotated secrets
    rotation_pattern = re.compile(r'(.+)_(\d{8})$')
    
    # Current timestamp
    current_time = datetime.now()
    retention_days = 7
    
    for keystore in keystores:
        match = rotation_pattern.match(keystore['name'])
        if match:
            secret_name = match.group(1)
            timestamp = match.group(2)
            
            # Parse timestamp
            secret_date = datetime.strptime(timestamp, '%Y%m%d')
            
            # Check if older than retention period
            if (current_time - secret_date).days > retention_days:
                print(f"Removing old secret: {keystore['name']}")
                trmm.delete_keystore(keystore['id'])
    
    print("Secret cleanup completed")
  EOT
  
  default_timeout = 60
}
```

## State Management

### Import Existing Keystore Entries

```bash
# Import by keystore ID
terraform import tacticalrmm_keystore.example 789
```

### State Characteristics

1. **Sensitivity Handling**: Values marked as sensitive in state
2. **Change Detection**: Value changes trigger updates
3. **Immutable Names**: Name changes require recreation

## Implementation Patterns

### 1. Hierarchical Key Naming

```hcl
locals {
  key_hierarchy = {
    credentials = {
      database  = ["connection_string", "username", "password"]
      api       = ["endpoint", "key", "secret"]
      storage   = ["account", "container", "sas_token"]
    }
    configuration = {
      features  = ["enabled_modules", "beta_flags", "limits"]
      network   = ["proxy_url", "dns_servers", "timeout"]
      logging   = ["level", "destination", "format"]
    }
  }
}

# Generate hierarchical keystore entries
resource "tacticalrmm_keystore" "hierarchical" {
  for_each = merge([
    for category, subcategories in local.key_hierarchy : {
      for subcat, keys in subcategories : {
        for key in keys : "${category}_${subcat}_${key}" => {
          name  = "${category}_${subcat}_${key}"
          value = var.keystore_values["${category}_${subcat}_${key}"]
        }
      }
    }
  ]...)
  
  name  = each.value.name
  value = each.value.value
}
```

### 2. Encryption Key Management

```hcl
# Master encryption key pattern
resource "tacticalrmm_keystore" "master_key" {
  name  = "master_encryption_key"
  value = var.master_key
}

# Derived keys for specific purposes
resource "tacticalrmm_keystore" "derived_keys" {
  for_each = toset(["database", "files", "communications"])
  
  name  = "${each.key}_encryption_key"
  value = bcrypt(format("%s:%s", var.master_key, each.key))
}
```

### 3. Configuration Templates

```hcl
# Store configuration templates
resource "tacticalrmm_keystore" "config_template" {
  name = "app_config_template"
  value = jsonencode({
    database = {
      host     = "{{DB_HOST}}"
      port     = "{{DB_PORT}}"
      name     = "{{DB_NAME}}"
      user     = "{{DB_USER}}"
      password = "{{DB_PASSWORD}}"
    }
    cache = {
      provider = "redis"
      host     = "{{CACHE_HOST}}"
      port     = 6379
      ttl      = 3600
    }
    logging = {
      level       = "{{LOG_LEVEL}}"
      destination = "{{LOG_DESTINATION}}"
    }
  })
}
```

## Best Practices

### 1. Security Considerations

- **Principle of Least Privilege**: Limit keystore access by role
- **Rotation Schedule**: Implement regular secret rotation
- **Audit Compliance**: Monitor keystore access patterns
- **Encryption Standards**: Use strong encryption for sensitive values

### 2. Naming Conventions

```hcl
# Standardized naming pattern
locals {
  naming_standard = {
    prefix      = var.organization_prefix
    environment = var.environment
    separator   = "_"
  }
  
  keystore_name = "${local.naming_standard.prefix}${local.naming_standard.separator}${local.naming_standard.environment}${local.naming_standard.separator}%s"
}
```

### 3. Value Validation

```hcl
# Validate values before storage
resource "tacticalrmm_keystore" "validated_entry" {
  name = "validated_config"
  value = jsonencode({
    validated = true
    timestamp = timestamp()
    checksum  = md5(var.config_content)
    content   = var.config_content
  })
  
  lifecycle {
    precondition {
      condition     = can(jsondecode(var.config_content))
      error_message = "Configuration must be valid JSON"
    }
  }
}
```

## Troubleshooting

### Common Issues

1. **Name Length Violations**
   - Maximum 25 characters enforced
   - Use abbreviations for long identifiers
   - Implement consistent naming scheme

2. **Value Size Limitations**
   - Large values may impact performance
   - Consider file storage for large data
   - Compress data when appropriate

3. **Access Permission Errors**
   - Verify API key permissions
   - Check keystore access policies
   - Review audit logs for denied access

### Diagnostic Patterns

```hcl
# Diagnostic keystore entries
resource "tacticalrmm_keystore" "diagnostics" {
  for_each = {
    last_update     = timestamp()
    provider_version = "0.1.0"
    terraform_version = terraform.version
    workspace       = terraform.workspace
  }
  
  name  = "diag_${each.key}"
  value = each.value
}
```

## Related Resources

- [tacticalrmm_script](script.md) - Scripts accessing keystore values
- [Data Source: tacticalrmm_keystore](../data-sources/keystore.md) - Query keystore entries
- [Data Source: tacticalrmm_keystores](../data-sources/keystores.md) - List all entries

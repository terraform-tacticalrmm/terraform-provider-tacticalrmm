# Tactical RMM Provider Documentation

Comprehensive documentation for the Terraform Provider for Tactical RMM, enabling infrastructure-as-code management of monitoring and automation resources.

## Provider Architecture

The Tactical RMM provider implements a systematic approach to resource management:

- **Resource Implementation**: Full CRUD lifecycle with precise state management
- **Authentication**: API key-based authentication with secure credential handling
- **Error Handling**: Comprehensive validation and error reporting
- **Import Support**: Native state import functionality for all resources

## Documentation Structure

### Provider Configuration
- [Provider Configuration](provider.md) - Authentication and setup parameters

### Resources
- [tacticalrmm_script](resources/script.md) - Automation script management
- [tacticalrmm_script_snippet](resources/script_snippet.md) - Reusable code snippet management
- [tacticalrmm_keystore](resources/keystore.md) - Secure key-value storage

### Data Sources
- [tacticalrmm_script](data-sources/script.md) - Query individual scripts
- [tacticalrmm_scripts](data-sources/scripts.md) - List all scripts
- [tacticalrmm_script_snippet](data-sources/script_snippet.md) - Query individual snippets
- [tacticalrmm_script_snippets](data-sources/script_snippets.md) - List all snippets
- [tacticalrmm_keystore](data-sources/keystore.md) - Query individual keystore entries
- [tacticalrmm_keystores](data-sources/keystores.md) - List all keystore entries

## Implementation Patterns

### Resource Design Principles

1. **State Management**
   - Computed fields automatically populated from API responses
   - Optional fields preserve null vs. empty distinctions
   - Array handling maintains user intent

2. **Error Handling**
   - HTTP status code validation
   - Detailed error messages with context
   - Graceful handling of missing resources

3. **Import Functionality**
   - All resources support `terraform import`
   - ID-based import for straightforward state recovery

### API Integration

The provider interfaces with Tactical RMM's Django REST API:
- RESTful endpoint mapping
- JSON request/response handling
- Consistent authentication headers

## Quick Start Guide

```hcl
# Configure the provider
provider "tacticalrmm" {
  endpoint = "https://api.your-trmm-instance.com"
  api_key  = var.trmm_api_key
}

# Create a maintenance script
resource "tacticalrmm_script" "disk_cleanup" {
  name        = "Disk Cleanup"
  description = "Automated disk cleanup script"
  shell       = "powershell"
  category    = "Maintenance"
  
  script_body = <<-EOT
    Remove-Item -Path "$env:TEMP\*" -Force -Recurse -ErrorAction SilentlyContinue
    Write-Output "Cleanup completed"
  EOT
  
  default_timeout = 300
}
```

## Development Guidelines

When extending the provider:

1. **Follow Established Patterns**: Review existing resource implementations
2. **Maintain Consistency**: Use the same error handling and state management approaches
3. **Document Thoroughly**: Include examples and edge case handling
4. **Test Comprehensively**: Unit and acceptance tests required

## Support

- GitHub Issues: Report bugs and feature requests
- Pull Requests: Contribute improvements following our guidelines
- Tactical RMM Documentation: Reference upstream API documentation

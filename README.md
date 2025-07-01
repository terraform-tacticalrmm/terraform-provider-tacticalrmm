# Terraform Provider for Tactical RMM

The Terraform Provider for Tactical RMM enables infrastructure-as-code management of Tactical RMM resources, providing systematic control over scripts, policies, checks, and monitoring configurations.

## Provider Overview

This provider interfaces with the Tactical RMM REST API to manage:
- **Scripts & Automation**: PowerShell, Python, Shell, and Batch scripts
- **Configuration Management**: Key-value storage, script snippets
- **Monitoring Resources**: Checks, tasks, and policies (upcoming)
- **Infrastructure Components**: Clients, sites, and agents (upcoming)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for development)
- Tactical RMM instance with API access
- Valid API key with appropriate permissions

## Installation

### Terraform Registry

```hcl
terraform {
  required_providers {
    tacticalrmm = {
      source  = "terraform-tacticalrmm/tacticalrmm"
      version = "~> 0.1.0"
    }
  }
}
```

### Local Development Build

```bash
git clone https://github.com/terraform-tacticalrmm/terraform-provider-tacticalrmm.git
cd terraform-provider-tacticalrmm
make install
```

## Provider Configuration

### Authentication Parameters

| Parameter | Type | Description | Environment Variable |
|-----------|------|-------------|---------------------|
| `endpoint` | String | Tactical RMM API endpoint URL | `TRMM_ENDPOINT` |
| `api_key` | String | API authentication key | `TRMM_API_KEY` |

### Configuration Example

```hcl
provider "tacticalrmm" {
  endpoint = "https://api.your-trmm-instance.com"
  api_key  = var.trmm_api_key
}
```

## Resource Coverage

### Currently Implemented

| Resource | Description | Status |
|----------|-------------|--------|
| `tacticalrmm_script` | Automation scripts management | ✅ Stable |
| `tacticalrmm_script_snippet` | Reusable code snippets | ✅ Stable |
| `tacticalrmm_keystore` | Secure key-value storage | ✅ Stable |

### Planned Implementation

| Resource | Description | Target Release |
|----------|-------------|----------------|
| `tacticalrmm_client` | Client organization management | v0.2.0 |
| `tacticalrmm_site` | Site/location management | v0.2.0 |
| `tacticalrmm_agent` | Agent deployment and configuration | v0.3.0 |
| `tacticalrmm_check` | Monitoring checks | v0.3.0 |
| `tacticalrmm_task` | Scheduled task automation | v0.3.0 |
| `tacticalrmm_policy` | Automation policy management | v0.4.0 |
| `tacticalrmm_alert_template` | Alert notification templates | v0.4.0 |

## Usage Examples

### Script Management

```hcl
resource "tacticalrmm_script" "maintenance" {
  name        = "System Maintenance"
  description = "Performs routine system maintenance"
  shell       = "powershell"
  category    = "Maintenance"
  
  script_body = file("${path.module}/scripts/maintenance.ps1")
  
  default_timeout     = 300
  run_as_user        = false
  supported_platforms = ["windows"]
}
```

### Script Snippet Creation

```hcl
resource "tacticalrmm_script_snippet" "common_functions" {
  name  = "CommonFunctions"
  desc  = "Shared PowerShell functions"
  shell = "powershell"
  code  = file("${path.module}/snippets/common.ps1")
}
```

### Secure Configuration Storage

```hcl
resource "tacticalrmm_keystore" "api_credentials" {
  name  = "external_api_key"
  value = var.external_api_key
}
```

## Data Sources

The provider includes comprehensive data sources for querying existing resources:

- `tacticalrmm_script` - Single script lookup
- `tacticalrmm_scripts` - List all scripts
- `tacticalrmm_script_snippet` - Single snippet lookup
- `tacticalrmm_script_snippets` - List all snippets
- `tacticalrmm_keystore` - Single keystore entry lookup
- `tacticalrmm_keystores` - List all keystore entries

## Development

### Architecture Overview

The provider follows a systematic implementation pattern:

1. **Resource Structure**: Each resource implements full CRUD operations
2. **State Management**: Precise handling of computed vs. user-defined attributes
3. **Error Handling**: Comprehensive API response validation
4. **Import Support**: All resources support state import functionality

### Building from Source

```bash
# Clone repository
git clone https://github.com/terraform-tacticalrmm/terraform-provider-tacticalrmm.git
cd terraform-provider-tacticalrmm

# Build provider
go build -o terraform-provider-tacticalrmm

# Run tests
go test ./...

# Install locally
make install
```

### Testing

```bash
# Unit tests
make test

# Acceptance tests (requires TRMM instance)
export TRMM_ENDPOINT="https://api.your-trmm-instance.com"
export TRMM_API_KEY="your-api-key"
make testacc
```

## Contributing

We welcome contributions following our systematic development approach:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/new-resource`)
3. Implement changes following existing patterns
4. Add comprehensive tests
5. Submit a pull request with detailed description

### Contribution Guidelines

- Maintain consistency with existing code patterns
- Include unit and acceptance tests
- Update documentation for new features
- Follow Go best practices and conventions

## License

This provider is distributed under the [MIT License](LICENSE).

## Support

- **Issues**: [GitHub Issues](https://github.com/terraform-tacticalrmm/terraform-provider-tacticalrmm/issues)
- **Discussions**: [GitHub Discussions](https://github.com/terraform-tacticalrmm/terraform-provider-tacticalrmm/discussions)
- **Tactical RMM**: [Official Documentation](https://docs.tacticalrmm.com/)

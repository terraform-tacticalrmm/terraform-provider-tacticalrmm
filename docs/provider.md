# Tactical RMM Provider Configuration

## Overview

The Tactical RMM provider establishes authenticated communication with your Tactical RMM instance, enabling Terraform to manage monitoring and automation resources through the REST API.

## Provider Configuration Parameters

### Required Parameters

None. All parameters can be supplied via environment variables.

### Optional Parameters

| Parameter | Type | Description | Environment Variable | Default |
|-----------|------|-------------|---------------------|---------|
| `endpoint` | String | Tactical RMM API endpoint URL | `TRMM_ENDPOINT` | `https://api.tactical-rmm.com` |
| `api_key` | String | API authentication key | `TRMM_API_KEY` | - |

## Authentication Methods

### Method 1: Direct Configuration

```hcl
provider "tacticalrmm" {
  endpoint = "https://api.your-trmm-instance.com"
  api_key  = "your-api-key-here"
}
```

### Method 2: Environment Variables

```bash
export TRMM_ENDPOINT="https://api.your-trmm-instance.com"
export TRMM_API_KEY="your-api-key-here"
```

```hcl
provider "tacticalrmm" {
  # Configuration loaded from environment
}
```

### Method 3: Variable-Based Configuration (Recommended)

```hcl
variable "trmm_api_key" {
  description = "Tactical RMM API Key"
  type        = string
  sensitive   = true
}

variable "trmm_endpoint" {
  description = "Tactical RMM API Endpoint"
  type        = string
  default     = "https://api.your-trmm-instance.com"
}

provider "tacticalrmm" {
  endpoint = var.trmm_endpoint
  api_key  = var.trmm_api_key
}
```

## API Key Generation

To generate an API key in Tactical RMM:

1. Log into your Tactical RMM instance as an administrator
2. Navigate to **Settings** → **Global Settings** → **API Keys**
3. Click **Add API Key**
4. Configure the following:
   - **Name**: Descriptive identifier (e.g., "Terraform Provider")
   - **Permissions**: Ensure appropriate permissions for resources you'll manage
   - **Expiration**: Set according to your security policies
5. Copy the generated API key immediately (it won't be shown again)

## Security Considerations

### Credential Protection

1. **Never commit API keys to version control**
   ```hcl
   # terraform.tfvars - ADD TO .gitignore
   trmm_api_key = "your-actual-api-key"
   ```

2. **Use environment variables in CI/CD pipelines**
   ```yaml
   env:
     TRMM_API_KEY: ${{ secrets.TRMM_API_KEY }}
   ```

3. **Leverage Terraform Cloud/Enterprise variable sets** for team environments

### Network Security

- Ensure HTTPS connectivity to your Tactical RMM instance
- Configure firewall rules to allow Terraform client access
- Consider using VPN or private networking for enhanced security

## Provider Initialization

### Standard Initialization

```bash
terraform init
```

### Development/Local Provider

```bash
# For local development builds
terraform init -plugin-dir=~/.terraform.d/plugins
```

## Configuration Examples

### Multi-Environment Setup

```hcl
# environments/prod/provider.tf
provider "trmm" {
  endpoint = "https://api.prod.trmm.company.com"
  api_key  = var.trmm_api_key_prod
}

# environments/dev/provider.tf
provider "trmm" {
  endpoint = "https://api.dev.trmm.company.com"
  api_key  = var.trmm_api_key_dev
}
```

### Provider Aliasing

```hcl
provider "trmm" {
  alias    = "primary"
  endpoint = "https://api.primary.trmm.company.com"
  api_key  = var.primary_api_key
}

provider "trmm" {
  alias    = "secondary"
  endpoint = "https://api.secondary.trmm.company.com"
  api_key  = var.secondary_api_key
}

# Use with resources
resource "tacticalrmm_script" "example" {
  provider = tacticalrmm.primary
  # ... resource configuration
}
```

## Troubleshooting

### Common Configuration Issues

1. **Authentication Failures**
   ```
   Error: Unable to authenticate with Tactical RMM API
   ```
   - Verify API key validity
   - Check API key permissions
   - Ensure endpoint URL is correct

2. **Connection Issues**
   ```
   Error: Unable to connect to Tactical RMM endpoint
   ```
   - Verify network connectivity
   - Check firewall rules
   - Validate HTTPS certificate

3. **Missing Configuration**
   ```
   Error: Missing API Key
   ```
   - Ensure `api_key` is provided via configuration or environment variable
   - Check environment variable naming (must be `TRMM_API_KEY`)

### Debug Mode

Enable detailed logging for troubleshooting:

```bash
export TF_LOG=DEBUG
terraform apply
```

## Provider Metadata

The provider automatically includes version information in API requests for compatibility tracking:

```
User-Agent: terraform-provider-tacticalrmm/0.1.0
```

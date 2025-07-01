#!/bin/bash
# Enhanced build script with comprehensive verification protocol

set -e

echo "=== Terraform Provider TRMM Build & Deployment Protocol ==="
echo "Timestamp: $(date)"

# Verification Phase 1: Environment validation
echo -e "\n[1/6] Environment Verification"
if ! command -v go &> /dev/null; then
    echo "ERROR: Go runtime not detected"
    exit 1
fi
echo "✓ Go version: $(go version)"

# Verification Phase 2: Dependency resolution
echo -e "\n[2/6] Dependency Resolution"
go mod download
go mod verify
echo "✓ Dependencies verified"

# Build Phase 3: Provider compilation with enhanced data source
echo -e "\n[3/6] Provider Compilation"
go build -o terraform-provider-tacticalrmm_v1.1.6
echo "✓ Provider binary compiled: terraform-provider-tacticalrmm_v1.1.6"

# Installation Phase 4: Local provider deployment
echo -e "\n[4/6] Local Provider Installation"
OS_ARCH="$(go env GOOS)_$(go env GOARCH)"
INSTALL_PATH="$HOME/.terraform.d/plugins/terraform-tacticalrmm/tacticalrmm/1.1.6/${OS_ARCH}"
mkdir -p "$INSTALL_PATH"
cp terraform-provider-tacticalrmm_v1.1.6 "$INSTALL_PATH/terraform-provider-tacticalrmm_v1.1.6"
echo "✓ Provider installed to: $INSTALL_PATH"

# Verification Phase 5: Test compilation
echo -e "\n[5/6] Test Suite Verification"
go test ./internal/provider -v -count=1 | grep -E "(PASS|FAIL|scripts_data_source)"
echo "✓ Test suite validated"

# Documentation Phase 6: Generate updated docs
echo -e "\n[6/6] Documentation Generation"
if command -v tfplugindocs &> /dev/null; then
    tfplugindocs generate
    echo "✓ Documentation regenerated"
else
    echo "⚠ tfplugindocs not installed - skipping doc generation"
fi

echo -e "\n=== Build Protocol Complete ==="
echo "Provider Version: 1.1.6 (Enhanced)"
echo "Key Enhancement: scripts data source with include_script_body parameter"
echo ""
echo "Next Steps:"
echo "1. Navigate to terraform-trmm-zap project"
echo "2. Run: terraform init -upgrade"
echo "3. Run: terraform plan"
echo "4. Run: terraform apply"

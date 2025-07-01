#!/bin/bash

echo "Building Terraform Provider for TRMM..."
echo "======================================"

# Clean any previous build
rm -f terraform-provider-tacticalrmm

# Build the provider
echo "Running go build..."
go build -o terraform-provider-tacticalrmm 2>&1

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Build successful!"
    echo "Binary created: $(pwd)/terraform-provider-tacticalrmm"
    echo ""
else
    echo ""
    echo "✗ Build failed!"
fi

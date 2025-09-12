#!/usr/bin/env bash
#
# Shiden Ledger Quick Start Script
# This script runs the introduction and installs Hyperledger Fabric
#

set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "Starting Shiden Ledger Quick Start..."
echo "Project root: ${PROJECT_ROOT}"
echo ""

# Step 1: Run the introduction script
echo "Step 1: Running introduction..."
if [ -f "${SCRIPT_DIR}/introduce.sh" ]; then
    chmod +x "${SCRIPT_DIR}/introduce.sh"
    "${SCRIPT_DIR}/introduce.sh"
else
    echo "Error: introduce.sh not found!"
    exit 1
fi

# Step 2: Install Hyperledger Fabric with specified version
echo "Step 2: Installing Hyperledger Fabric v2.5.13..."
if [ -f "${SCRIPT_DIR}/install-fabric.sh" ]; then
    chmod +x "${SCRIPT_DIR}/install-fabric.sh"
    cd "${PROJECT_ROOT}"
    "${SCRIPT_DIR}/install-fabric.sh" --fabric-version 2.5.13
    echo ""
    echo "✅ Hyperledger Fabric installation completed!"
else
    echo "Error: install-fabric.sh not found!"
    exit 1
fi

echo ""
echo "🎉 Shiden Ledger quick start completed successfully!"
echo "Your development environment is now ready."
echo ""
echo "Next steps:"
echo "- Navigate to fabric-samples directory to explore examples"
echo "- Check the test-network for a quick blockchain setup"
echo "- Review the documentation for advanced configuration"
echo ""
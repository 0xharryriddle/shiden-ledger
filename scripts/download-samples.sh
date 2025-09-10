#!/usr/bin/env bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Script to download Hyperledger Fabric samples with user-specified version
# This script prompts the user to input the desired version from keyboard

set -e

# Function to display help information
print_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Download Hyperledger Fabric samples and binaries with interactive version selection"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -v, --version       Specify version directly (skips interactive input)"
    echo "  -s, --samples-only  Only download samples, don't run install-fabric.sh"
    echo "  --save-only         Only save version to file, don't download anything"
    echo ""
    echo "By default, this script will:"
    echo "1. Ask for the Hyperledger Fabric version"
    echo "2. Download fabric-samples for that version"
    echo "3. Run install-fabric.sh to download binaries and docker images"
    echo ""
    echo "The selected version will be saved to .fabric_version for later use."
}

# Function to validate version format
validate_version() {
    local version=$1
    # Check if version matches pattern like 2.5.12, 1.4.7, etc.
    if [[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        return 0
    else
        return 1
    fi
}

# Function to prompt for version input
prompt_for_version() {
    echo "Please enter the Hyperledger Fabric version you want to download:"
    echo "Examples: 2.5.12, 2.4.9, 1.4.12"
    echo -n "Version: "
    read -r VERSION
    
    # Validate the input
    if ! validate_version "$VERSION"; then
        echo "Error: Invalid version format. Please use format like 2.5.12"
        echo "Trying again..."
        echo ""
        prompt_for_version
    fi
}

# Function to save version to file
save_version() {
    local version_file=".fabric_version"
    echo "FABRIC_VERSION=${VERSION}" > "$version_file"
    echo "===> Fabric version ${VERSION} saved to ${version_file}"
    echo "You can now use: source ${version_file} && ./install-fabric.sh -f \$FABRIC_VERSION samples"
}

# Function to run install-fabric.sh with the selected version
run_install_fabric() {
    echo ""
    echo "===> Running install-fabric.sh to download binaries and docker images for version ${VERSION}"
    if [ -f "./install-fabric.sh" ]; then
        ./install-fabric.sh -f "${VERSION}" samples binary docker
    else
        echo "Error: install-fabric.sh not found in current directory"
        echo "Please run from the repository root directory."
        return 1
    fi
}
cloneSamplesRepo() {
    echo "===> Downloading fabric-samples version ${VERSION}"
    
    # Check if we're already in fabric-samples repo
    if [ -d test-network ]; then
        echo "==> Already in fabric-samples repo"
    elif [ -d fabric-samples ]; then
        echo "===> Changing directory to fabric-samples"
        cd fabric-samples
    else
        echo "===> Cloning hyperledger/fabric-samples repo"
        git clone -b main https://github.com/hyperledger/fabric-samples.git && cd fabric-samples
    fi

    # Try to checkout the specified version
    if GIT_DIR=.git git rev-parse v${VERSION} >/dev/null 2>&1; then
        echo "===> Checking out v${VERSION} of hyperledger/fabric-samples"
        git checkout -q v${VERSION}
        echo "===> Successfully checked out fabric-samples v${VERSION}"
    else
        echo "Warning: fabric-samples v${VERSION} does not exist, defaulting to main branch."
        echo "The main branch is intended to work with recent versions of fabric."
        git checkout -q main
    fi
    
    echo "===> Fabric samples downloaded successfully!"
    echo "Location: $(pwd)"
}

# Parse command line arguments
VERSION=""
SAMPLES_ONLY=false
SAVE_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            print_help
            exit 0
            ;;
        -v|--version)
            if [ -n "$2" ] && [[ $2 != -* ]]; then
                VERSION="$2"
                shift 2
            else
                echo "Error: --version requires a value"
                exit 1
            fi
            ;;
        -s|--samples-only)
            SAMPLES_ONLY=true
            shift
            ;;
        --save-only)
            SAVE_ONLY=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            print_help
            exit 1
            ;;
    esac
done

# Main execution
echo "Hyperledger Fabric Complete Setup Script"
echo "========================================"
echo ""

# If version not provided via command line, prompt for it
if [ -z "$VERSION" ]; then
    prompt_for_version
else
    if ! validate_version "$VERSION"; then
        echo "Error: Invalid version format provided: $VERSION"
        echo "Please use format like 2.5.12"
        exit 1
    fi
fi

echo ""
echo "Selected version: $VERSION"
echo ""

# Always save the version
save_version

# If save-only mode, exit here
if [ "$SAVE_ONLY" = true ]; then
    echo "Version saved. Exiting as requested (--save-only mode)."
    exit 0
fi

# Show what will be downloaded
if [ "$SAMPLES_ONLY" = true ]; then
    echo "This will download:"
    echo "- fabric-samples repository (version ${VERSION})"
else
    echo "This will download:"
    echo "- fabric-samples repository (version ${VERSION})"
    echo "- Hyperledger Fabric binaries (version ${VERSION})"
    echo "- Docker images for Fabric and Fabric CA"
fi
echo ""

# Confirm with user
echo -n "Do you want to proceed? (y/N): "
read -r CONFIRM

if [[ $CONFIRM =~ ^[Yy]$ ]]; then
    echo ""
    echo "===> Starting download process..."
    
    # Always download samples first
    cloneSamplesRepo
    
    # Run install-fabric.sh unless samples-only mode
    if [ "$SAMPLES_ONLY" = false ]; then
        run_install_fabric
        echo ""
        echo "===> Complete! All components downloaded successfully."
        echo "===> Fabric version ${VERSION} is ready to use."
    else
        echo ""
        echo "===> Samples download complete!"
        echo "===> To download binaries and docker images later, run:"
        echo "     source .fabric_version && ./install-fabric.sh -f \$FABRIC_VERSION binary docker"
    fi
else
    echo "Download cancelled."
    exit 0
fi
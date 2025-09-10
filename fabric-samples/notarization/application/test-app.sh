#!/bin/bash

# Notarization Application Gateway Test Script
# This script helps test the connection to your Fabric network

echo "🌟 Notarization Network Test Script"
echo "=================================="

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found!"
    echo "Please create a .env file with the following variables:"
    echo "FABRIC_PEER_ENDPOINT=localhost:7051"
    echo "FABRIC_TLS_CERT_PATH=./crypto/peer/tls/ca.crt"
    echo "FABRIC_MSP_ID=VPCC1MSP"
    echo "FABRIC_CERT_PATH=./wallet/user/cert.pem"
    echo "FABRIC_KEY_PATH=./wallet/user/key_sk.pem"
    echo "FABRIC_CHANNEL=notary-main"
    echo "FABRIC_CHAINCODE=notarycc"
    echo "FABRIC_CONTRACT=NotarizationContract"
    exit 1
fi

# Source the environment variables
source .env

echo "🔧 Configuration:"
echo "  Peer: $FABRIC_PEER_ENDPOINT"
echo "  MSP ID: $FABRIC_MSP_ID"
echo "  Channel: $FABRIC_CHANNEL"
echo "  Chaincode: $FABRIC_CHAINCODE"
echo "  Contract: $FABRIC_CONTRACT"
echo ""

# Check if required files exist
echo "📁 Checking required files..."
files_to_check=(
    "$FABRIC_TLS_CERT_PATH"
    "$FABRIC_CERT_PATH"
    "$FABRIC_KEY_PATH"
)

for file in "${files_to_check[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✅ $file"
    else
        echo "  ❌ $file (missing)"
        MISSING_FILES=1
    fi
done

if [ "$MISSING_FILES" = "1" ]; then
    echo ""
    echo "❌ Some required files are missing. Please ensure your crypto material is in place."
    exit 1
fi

echo ""
echo "🚀 Running notarization application..."
echo "====================================="

# Run the application
./notarization-app

echo ""
echo "✨ Test completed!"

# Notarization Application Gateway

This application gateway provides a Go-based client interface to interact with the Hyperledger Fabric notarization network. It demonstrates how to connect to your fabric network and perform notarization operations.

## Features

- 🔐 **Secure Connection**: TLS-enabled connection to Fabric peers
- 📋 **Instrument Management**: Issue, retrieve, verify, and revoke notarization instruments
- 🔒 **Private Data**: Store PII data in private data collections
- 📡 **Event Listening**: Real-time monitoring of chaincode events
- ✅ **Error Handling**: Comprehensive error handling and logging
- 🎯 **Type Safety**: Strongly typed Go structures for all operations

## Prerequisites

1. **Hyperledger Fabric Network**: A running Fabric network with the notarization chaincode deployed
2. **Crypto Material**: Valid certificates and private keys for your organization
3. **Go 1.22+**: Go programming language installed
4. **Network Access**: Connectivity to the Fabric peer endpoints

## Configuration

Create a `.env` file in the application directory with the following variables:

```env
# Fabric Network Configuration
FABRIC_PEER_ENDPOINT=localhost:7051
FABRIC_TLS_CERT_PATH=./crypto/peer/tls/ca.crt
FABRIC_MSP_ID=VPCC1MSP
FABRIC_CERT_PATH=./wallet/user/cert.pem
FABRIC_KEY_PATH=./wallet/user/key_sk.pem

# Chaincode Configuration
FABRIC_CHANNEL=notary-main
FABRIC_CHAINCODE=notarycc
FABRIC_CONTRACT=NotarizationContract
```

## Directory Structure

```
application/
├── .env                    # Environment configuration
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── notarization-app        # Compiled application binary
├── test-app.sh            # Test script
├── application-gateway/
│   └── notarization.go    # Main application code
├── crypto/                # Crypto material (you need to provide this)
│   └── peer/
│       └── tls/
│           └── ca.crt
└── wallet/                # User certificates (you need to provide this)
    └── user/
        ├── cert.pem
        └── key_sk.pem
```

## Building and Running

### Build the Application

```bash
cd fabric-samples/notarization/application
go build -o notarization-app ./application-gateway
```

### Run the Application

1. **Using the test script** (recommended):

   ```bash
   ./test-app.sh
   ```

2. **Direct execution**:
   ```bash
   source .env
   ./notarization-app
   ```

## API Operations

The application demonstrates the following notarization operations:

### 1. Issue Instrument

Creates a new notarization instrument with digital signatures and metadata.

### 2. Store PII Data

Securely stores personally identifiable information in private data collections.

### 3. Retrieve Instrument

Fetches an existing instrument by its unique identifier.

### 4. Verify Instrument

Validates an instrument against a document hash to ensure integrity.

### 5. Revoke Instrument (Optional)

Revokes an existing instrument with a specified reason.

## Code Structure

### Main Components

- **`NotarizationClient`**: Main client wrapper for Fabric Gateway
- **`Config`**: Configuration structure loaded from environment
- **`ExampleInstrumentPayload`**: Typed structure for instrument data
- **Event Handling**: Real-time chaincode event monitoring

### Key Methods

```go
// Create new client
client, err := NewNotarizationClient(cfg)

// Issue instrument
result, err := client.InstrumentIssue(payloadJSON, requireMOJ)

// Store PII data
err := client.PutPII(caseID, piiJSON)

// Get instrument
data, err := client.InstrumentGet(instrumentID)

// Verify instrument
result, err := client.InstrumentVerify(instrumentID, docHash)

// Revoke instrument
result, err := client.InstrumentRevoke(instrumentID, reason)
```

## Event Monitoring

The application automatically subscribes to chaincode events and logs them in real-time:

```
📢 Event: instrument.issued | TxID: abc123... | Payload: {...}
📢 Event: instrument.revoked | TxID: def456... | Payload: {...}
```

## Error Handling

The application includes comprehensive error handling with descriptive messages:

- Connection errors to Fabric network
- Certificate and key validation errors
- Transaction endorsement and commit failures
- Chaincode execution errors

## Sample Output

```
🌟 Notarization Application Gateway Starting...
🔧 Configuration loaded:
   Peer: localhost:7051
   MSP ID: VPCC1MSP
   Channel: notary-main
   Chaincode: notarycc
   Contract: NotarizationContract
✅ Connected to Fabric network successfully!
Event listener started...

🚀 Starting Notarization Workflow Demonstration

📋 Step 1: Issuing Instrument
✅ Instrument issued successfully
📄 Issued instrument result: {...}

🔐 Step 2: Storing PII Data
✅ PII stored successfully for case: CASE-2025-0001

🔍 Step 3: Retrieving Instrument
✅ Retrieved instrument: INS-2025-0000001
📄 Retrieved instrument: {...}

✅ Step 4: Verifying Instrument
✅ Verification completed for instrument: INS-2025-0000001
🔐 Verification result: {...}

✨ Notarization workflow completed successfully!
🎉 Application completed successfully!
```

## Troubleshooting

### Common Issues

1. **Connection Failed**: Check peer endpoint and TLS certificate paths
2. **Authentication Failed**: Verify MSP ID and user certificates are correct
3. **Chaincode Not Found**: Ensure chaincode is deployed to the specified channel
4. **Permission Denied**: Check if user has proper roles for operations

### Debugging

Enable verbose logging by modifying the log level in the code or check Fabric peer logs for detailed error information.

## Security Considerations

- Store private keys securely and never commit them to version control
- Use proper TLS certificates for production environments
- Implement proper access controls for PII data
- Monitor and audit all notarization operations

## License

This application is part of the Hyperledger Fabric samples and follows the same license terms.

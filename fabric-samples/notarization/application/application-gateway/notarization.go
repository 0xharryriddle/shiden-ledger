package main

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NotarizationClient wraps the Fabric Gateway client with notarization-specific methods
type NotarizationClient struct {
	gateway    *client.Gateway
	contract   *client.Contract
	network    *client.Network
	conn       *grpc.ClientConn
	stopEvents func()
}

// Config holds configuration for connecting to the Fabric network
type Config struct {
	PeerEndpoint string
	TLSCertPath  string
	MSPID        string
	CertPath     string
	KeyPath      string
	Channel      string
	Chaincode    string
	Contract     string
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	cfg := &Config{
		PeerEndpoint: os.Getenv("FABRIC_PEER_ENDPOINT"),
		TLSCertPath:  os.Getenv("FABRIC_TLS_CERT_PATH"),
		MSPID:        os.Getenv("FABRIC_MSP_ID"),
		CertPath:     os.Getenv("FABRIC_CERT_PATH"),
		KeyPath:      os.Getenv("FABRIC_KEY_PATH"),
		Channel:      os.Getenv("FABRIC_CHANNEL"),
		Chaincode:    os.Getenv("FABRIC_CHAINCODE"),
		Contract:     os.Getenv("FABRIC_CONTRACT"),
	}

	// Validate required fields
	if cfg.PeerEndpoint == "" || cfg.TLSCertPath == "" || cfg.MSPID == "" ||
		cfg.CertPath == "" || cfg.KeyPath == "" || cfg.Channel == "" ||
		cfg.Chaincode == "" {
		return nil, errors.New("missing required environment variables")
	}

	return cfg, nil
}

func must(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func fileBytes(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read file %s: %v", path, err)
	}
	return b
}

func tlsCredentials(tlsCertPath string) (credentials.TransportCredentials, error) {
	pemBytes := fileBytes(tlsCertPath)
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(pemBytes) {
		return nil, errors.New("failed to add peer TLS cert")
	}
	return credentials.NewClientTLSFromCert(cp, ""), nil
}

// NewNotarizationClient creates a new client connected to the Fabric network
func NewNotarizationClient(cfg *Config) (*NotarizationClient, error) {
	// Create identity
	cert, err := identity.CertificateFromPEM(fileBytes(cfg.CertPath))
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	id, err := identity.NewX509Identity(cfg.MSPID, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	pk, err := identity.PrivateKeyFromPEM(fileBytes(cfg.KeyPath))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	sign, err := identity.NewPrivateKeySign(pk)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	// Create gRPC connection
	creds, err := tlsCredentials(cfg.TLSCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS credentials: %w", err)
	}

	conn, err := grpc.Dial(cfg.PeerEndpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Create gateway
	gw, err := client.Connect(id,
		client.WithClientConnection(conn),
		client.WithSign(sign),
		client.WithEvaluateTimeout(10*time.Second),
		client.WithEndorseTimeout(30*time.Second),
		client.WithSubmitTimeout(30*time.Second),
		client.WithCommitStatusTimeout(60*time.Second),
	)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	// Get network and contract
	network := gw.GetNetwork(cfg.Channel)
	var contract *client.Contract
	if cfg.Contract != "" {
		contract = network.GetContractWithName(cfg.Chaincode, cfg.Contract)
	} else {
		contract = network.GetContract(cfg.Chaincode)
	}

	// Start event listening
	stopEvents := startEventListener(network, cfg.Chaincode)

	return &NotarizationClient{
		gateway:    gw,
		contract:   contract,
		network:    network,
		conn:       conn,
		stopEvents: stopEvents,
	}, nil
}

// Close closes the client connection
func (nc *NotarizationClient) Close() {
	if nc.stopEvents != nil {
		nc.stopEvents()
	}
	if nc.gateway != nil {
		nc.gateway.Close()
	}
	if nc.conn != nil {
		nc.conn.Close()
	}
}

// startEventListener starts listening to chaincode events
func startEventListener(network *client.Network, chaincode string) func() {
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := network.ChaincodeEvents(ctx, chaincode)
	if err != nil {
		log.Printf("Warning: Failed to subscribe to events: %v", err)
		return cancel
	}

	go func() {
		log.Println("Event listener started...")
		for evt := range ch {
			log.Printf("📢 Event: %s | TxID: %s | Payload: %s",
				evt.EventName, evt.TransactionID, string(evt.Payload))
		}
		log.Println("Event listener stopped")
	}()

	return cancel
}

// PutPII stores personally identifiable information in the private data collection
func (nc *NotarizationClient) PutPII(caseID string, piiJSON []byte) error {
	log.Printf("🔐 Storing PII for case: %s", caseID)

	transient := map[string][]byte{"pii": piiJSON}
	proposal, err := nc.contract.NewProposal(
		"PutPII",
		client.WithArguments(caseID),
		client.WithTransient(transient),
	)
	if err != nil {
		return fmt.Errorf("failed to create PutPII proposal: %w", err)
	}

	txn, err := proposal.Endorse()
	if err != nil {
		return fmt.Errorf("failed to endorse PutPII: %w", err)
	}

	_, err = txn.Submit()
	if err != nil {
		return fmt.Errorf("failed to submit PutPII: %w", err)
	}

	log.Printf("✅ PII stored successfully for case: %s", caseID)
	return nil
}

// InstrumentIssue creates a new notarization instrument
func (nc *NotarizationClient) InstrumentIssue(payloadJSON []byte, requireMOJ bool) ([]byte, error) {
	argRequire := strconv.FormatBool(requireMOJ)

	log.Printf("📋 Issuing instrument (requireMOJ: %v)", requireMOJ)

	result, err := nc.contract.SubmitTransaction("InstrumentIssue", string(payloadJSON), argRequire)
	if err != nil {
		return nil, fmt.Errorf("failed to issue instrument: %w", err)
	}

	log.Printf("✅ Instrument issued successfully")
	return result, nil
}

// InstrumentGet retrieves an instrument by ID
func (nc *NotarizationClient) InstrumentGet(id string) ([]byte, error) {
	log.Printf("🔍 Getting instrument: %s", id)

	result, err := nc.contract.EvaluateTransaction("InstrumentGet", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get instrument %s: %w", id, err)
	}

	log.Printf("✅ Retrieved instrument: %s", id)
	return result, nil
}

// InstrumentVerify verifies an instrument against a document hash
func (nc *NotarizationClient) InstrumentVerify(id, docHash string) ([]byte, error) {
	log.Printf("🔐 Verifying instrument %s with hash: %s", id, docHash)

	result, err := nc.contract.EvaluateTransaction("InstrumentVerify", id, docHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify instrument %s: %w", id, err)
	}

	log.Printf("✅ Verification completed for instrument: %s", id)
	return result, nil
}

// InstrumentRevoke revokes an existing instrument
func (nc *NotarizationClient) InstrumentRevoke(id, reason string) ([]byte, error) {
	log.Printf("⚠️  Revoking instrument %s, reason: %s", id, reason)

	result, err := nc.contract.SubmitTransaction("InstrumentRevoke", id, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke instrument %s: %w", id, err)
	}

	log.Printf("✅ Instrument revoked: %s", id)
	return result, nil
}

// ExampleInstrumentPayload represents a sample instrument payload
type ExampleInstrumentPayload struct {
	ID           string      `json:"id"`
	CaseID       string      `json:"caseId"`
	InstrumentNo string      `json:"instrumentNo"`
	Province     string      `json:"province"`
	Parties      []PartyRef  `json:"parties"`
	DocHash      string      `json:"docHash"`
	OffchainURI  string      `json:"offchainUri"`
	JournalSeq   int         `json:"journalSeq"`
	QR           string      `json:"qr"`
	Signatures   []Signature `json:"signatures"`
}

type PartyRef struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	Name string `json:"name"`
}

type Signature struct {
	Subject string `json:"subject"`
	CertSN  string `json:"certSn"`
	Algo    string `json:"algo"`
	Time    string `json:"time"`
}

func demonstrateNotarizationWorkflow(client *NotarizationClient) error {
	log.Println("🚀 Starting Notarization Workflow Demonstration")

	// 1. Create example instrument payload
	payload := ExampleInstrumentPayload{
		ID:           "INS-2025-0000001",
		CaseID:       "CASE-2025-0001",
		InstrumentNo: "2025/VPCC1/0000001",
		Province:     "79",
		Parties: []PartyRef{
			{ID: "P1", Role: "SELLER", Name: "Nguyen A"},
		},
		DocHash:     "sha256:abcd1234567890efgh",
		OffchainURI: "s3://bucket/final.pdf",
		JournalSeq:  1234,
		QR:          "INS-2025-0000001",
		Signatures: []Signature{
			{
				Subject: "CCV Nguyen A",
				CertSN:  "12AB",
				Algo:    "RSA-PSS",
				Time:    time.Now().Format(time.RFC3339),
			},
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 2. Issue the instrument
	log.Println("\n📋 Step 1: Issuing Instrument")
	result, err := client.InstrumentIssue(payloadJSON, true)
	if err != nil {
		return fmt.Errorf("failed to issue instrument: %w", err)
	}
	log.Printf("📄 Issued instrument result: %s\n", string(result))

	// 3. Store PII data
	log.Println("🔐 Step 2: Storing PII Data")
	pii := map[string]interface{}{
		"hmacCccd":      "a1b2c3d4e5f6...",
		"maritalStatus": "married",
		"phoneNumber":   "+84901234567",
	}
	piiJSON, _ := json.Marshal(pii)

	if err := client.PutPII("CASE-2025-0001", piiJSON); err != nil {
		log.Printf("⚠️  Warning: Failed to store PII: %v", err)
		// Don't fail the demo for PII errors
	}

	// 4. Retrieve the instrument
	log.Println("\n🔍 Step 3: Retrieving Instrument")
	retrieved, err := client.InstrumentGet("INS-2025-0000001")
	if err != nil {
		return fmt.Errorf("failed to get instrument: %w", err)
	}
	log.Printf("📄 Retrieved instrument: %s\n", string(retrieved))

	// 5. Verify the instrument
	log.Println("✅ Step 4: Verifying Instrument")
	verification, err := client.InstrumentVerify("INS-2025-0000001", "sha256:abcd1234567890efgh")
	if err != nil {
		return fmt.Errorf("failed to verify instrument: %w", err)
	}
	log.Printf("🔐 Verification result: %s\n", string(verification))

	// 6. Optional: Demonstrate revocation (commented out to not affect the demo)
	/*
		log.Println("⚠️  Step 5: Revoking Instrument (Demo)")
		revoked, err := client.InstrumentRevoke("INS-2025-0000001", "Demo revocation")
		if err != nil {
			return fmt.Errorf("failed to revoke instrument: %w", err)
		}
		log.Printf("📄 Revoked instrument: %s\n", string(revoked))
	*/

	log.Println("\n✨ Notarization workflow completed successfully!")
	return nil
}

func main() {
	log.Println("🌟 Notarization Application Gateway Starting...")

	// Load configuration
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	log.Printf("🔧 Configuration loaded:")
	log.Printf("   Peer: %s", cfg.PeerEndpoint)
	log.Printf("   MSP ID: %s", cfg.MSPID)
	log.Printf("   Channel: %s", cfg.Channel)
	log.Printf("   Chaincode: %s", cfg.Chaincode)
	log.Printf("   Contract: %s", cfg.Contract)

	// Create notarization client
	client, err := NewNotarizationClient(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}
	defer client.Close()

	log.Println("✅ Connected to Fabric network successfully!")

	// Run demonstration workflow
	if err := demonstrateNotarizationWorkflow(client); err != nil {
		log.Fatalf("❌ Demo failed: %v", err)
	}

	log.Println("🎉 Application completed successfully!")
}

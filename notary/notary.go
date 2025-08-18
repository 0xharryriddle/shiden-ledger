package notary

import (
	"log"

	"github.com/0xharryriddle/shiden-ledger/notary/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

func main() {
	notaryChaincode, err := contractapi.NewChaincode(&chaincode.NotaryContract{})
	if err != nil {
		log.Panicf("Error creating notary chaincode: %v", err)
	}

	if err := notaryChaincode.Start(); err != nil {
		log.Panicf("Error starting notary chaincode: %v", err)
	}
}

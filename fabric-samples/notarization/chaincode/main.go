package contracts

import (
	notarization "github.com/0xharryriddle/shiden-ledger/fabric-samples/notarization/contracts/notarization"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	notarizationContract, err := contractapi.NewChaincode(&notarization.NotarizationContract{})
	if err != nil {
		panic("Error creating notarization chaincode: " + err.Error())
	}

	if err := notarizationContract.Start(); err != nil {
		panic("Error starting notarization chaincode: " + err.Error())
	}
}

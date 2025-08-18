package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

/* -------------------------------- Constants ------------------------------- */

type DocumentStatusType string

var DocumentStatus = struct {
	DRAFT   DocumentStatusType
	ISSUED  DocumentStatusType
	REVOKED DocumentStatusType
	EXPIRED DocumentStatusType
}{
	DRAFT:   "DRAFT",
	ISSUED:  "ISSUED",
	REVOKED: "REVOKED",
	EXPIRED: "EXPIRED",
}

/* ------------------------------ SmartContract ----------------------------- */

type NotaryContract struct {
	contractapi.Contract
}

/* --------------------------------- Events --------------------------------- */
type Events struct {
}

/* --------------------------------- Assets --------------------------------- */

type NotarizedDocument struct {
	DocID        string             `json:"docId"`
	DocType      string             `json:"docType"`
	TemplateRef  string             `json:"templateRef"`
	OwnerSubject string             `json:"ownerSubject"`
	ContentHash  string             `json:"contentHash"`
	Status       DocumentStatusType `json:"status"`
	Signatures   []string           `json:"signatures"`
}

type History struct {
	TxID      string `json:"txId"`
	Timestamp int64  `json:"timestamp"`
	IsDelete  bool   `json:"isDelete"`
	Status    string `json:"status"`
}

func (c *NotaryContract) CreateDocument(ctx contractapi.TransactionContextInterface, payload string) (string, error) {
	var d NotarizedDocument
	if err := json.Unmarshal([]byte(payload), &d); err != nil {
		return "", err
	}
	if d.DocID == "" || d.ContentHash == "" {
		return "", fmt.Errorf("missing docId or contentHash")
	}
	d.Status = "DRAFT"
	b, _ := json.Marshal(d)
	if err := ctx.GetStub().PutState(d.DocID, b); err != nil {
		return "", err
	}
	return d.DocID, nil
}

func (c *NotaryContract) EndorseDocument(ctx contractapi.TransactionContextInterface, docID string, signature string) error {
	b, err := ctx.GetStub().GetState(docID)
	if err != nil || b == nil {
		return fmt.Errorf("not found")
	}
	var d NotarizedDocument
	_ = json.Unmarshal(b, &d)
	d.Signatures = append(d.Signatures, signature)
	d.Status = "ISSUED"
	nb, _ := json.Marshal(d)
	return ctx.GetStub().PutState(docID, nb)
}

func (c *NotaryContract) GetDocument(ctx contractapi.TransactionContextInterface, docID string) (*NotarizedDocument, error) {
	b, err := ctx.GetStub().GetState(docID)
	if err != nil || b == nil {
		return nil, fmt.Errorf("not found")
	}
	var d NotarizedDocument
	_ = json.Unmarshal(b, &d)
	return &d, nil
}

func (c *NotaryContract) RevokeDocument(ctx contractapi.TransactionContextInterface, docID, reason string) error {
	b, err := ctx.GetStub().GetState(docID)
	if err != nil || b == nil {
		return fmt.Errorf("not found")
	}
	var d NotarizedDocument
	_ = json.Unmarshal(b, &d)
	d.Status = "REVOKED"
	nb, _ := json.Marshal(d)
	return ctx.GetStub().PutState(docID, nb)
}

func (c *NotaryContract) VerifyDocument(ctx contractapi.TransactionContextInterface, docID, contentHash string) (string, error) {
	b, err := ctx.GetStub().GetState(docID)
	if err != nil || b == nil {
		return "INVALID", nil
	}
	var d NotarizedDocument
	_ = json.Unmarshal(b, &d)
	if d.Status == "REVOKED" {
		return "REVOKED", nil
	}
	if d.ContentHash == contentHash {
		return "VALID", nil
	}
	return "INVALID", nil
}

func (c *NotaryContract) History(ctx contractapi.TransactionContextInterface, docID string) ([]*History, error) {
	it, err := ctx.GetStub().GetHistoryForKey(docID)
	if err != nil {
		return nil, err
	}
	defer it.Close()
	var out []*History
	for it.HasNext() {
		mod, err := it.Next()
		if err != nil {
			return nil, err
		}
		var d NotarizedDocument
		if len(mod.Value) > 0 {
			_ = json.Unmarshal(mod.Value, &d)
		}
		ts := int64(0)
		if mod.Timestamp != nil {
			ts = mod.Timestamp.AsTime().Unix()
		}
		out = append(out, &History{TxID: mod.TxId, Timestamp: ts, IsDelete: mod.IsDelete, Status: string(d.Status)})
	}
	return out, nil
}

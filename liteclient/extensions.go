package liteclient

//go:generate go run generator.go

import (
	"encoding/json"
	"fmt"
)

var (
	ErrBlockNotApplied = fmt.Errorf("block is not applied")
)

func (t LiteServerError) Error() string {
	return fmt.Sprintf("error code: %d message: %s", t.Code, t.Message)
}

func (t LiteServerError) IsNotApplied() bool {
	return t.Message == "block is not applied"
}

func (t LiteServerBlockLink) MarshalJSON() ([]byte, error) {
	switch t.SumType {
	case "LiteServerBlockLinkBack":
		return json.Marshal(struct {
			SumType                 string
			LiteServerBlockLinkBack struct {
				ToKeyBlock bool
				From       TonNodeBlockIdExt
				To         TonNodeBlockIdExt
				DestProof  []byte
				Proof      []byte
				StateProof []byte
			}
		}{SumType: "lite_server_block_link_back", LiteServerBlockLinkBack: t.LiteServerBlockLinkBack})
	case "LiteServerBlockLinkForward":
		return json.Marshal(struct {
			SumType                    string
			LiteServerBlockLinkForward struct {
				ToKeyBlock  bool
				From        TonNodeBlockIdExt
				To          TonNodeBlockIdExt
				DestProof   []byte
				ConfigProof []byte
				Signatures  LiteServerSignatureSet
			}
		}{SumType: "lite_server_block_link_forward", LiteServerBlockLinkForward: t.LiteServerBlockLinkForward})
	default:
		return nil, fmt.Errorf("invalid sumtype")
	}
}

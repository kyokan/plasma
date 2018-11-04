package pb

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"encoding/json"
	"fmt"
)

func (m *BigInt) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", m.Hex)), nil
}

func (m *Input) MarshalJSON() ([]byte, error) {
	if m.TxIdx == 0 && m.BlockNum == 0 && m.OutIdx == 0 {
		return []byte("null"), nil
	}

	return json.Marshal(m)
}

type rawOutput struct {
	NewOwner string  `json:"newOwner"`
	Amount   *BigInt `json:"amount"`
}

func (m Output) MarshalJSON() ([]byte, error) {
	if len(m.NewOwner) == 0 && m.Amount == nil {
		return []byte("\"null\""), nil
	}

	raw := &rawOutput{
		hexutil.Encode(m.NewOwner),
		m.Amount,
	}
	return json.Marshal(raw)
}

type rawBlockHeader struct {
	MerkleRoot    string  `json:"merkleRoot"`
	RLPMerkleRoot string  `json:"rlpMerkleRoot"`
	PrevHash      *string `json:"prevHash"`
	Number        uint64  `json:"number"`
}

func (m BlockHeader) MarshalJSON() ([]byte, error) {
	raw := &rawBlockHeader{
		MerkleRoot:    hexutil.Encode(m.MerkleRoot),
		RLPMerkleRoot: hexutil.Encode(m.RlpMerkleRoot),
		PrevHash:      hexOrNil(m.PrevHash),
		Number:        m.Number,
	}
	return json.Marshal(raw)
}

type rawTransaction struct {
	Input0   *Input  `json:"input0"`
	Sig0     *string `json:"sig0"`
	Input1   *Input  `json:"input1"`
	Sig1     *string `json:"sig1"`
	Output0  *Output `json:"output0"`
	Output1  *Output `json:"output1"`
	Fee      *BigInt `json:"fee"`
	BlockNum uint64  `json:"blockNum"`
	TxIdx    uint32  `json:"txIdx"`
	RootSig  string  `json:"rootSig"`
}

func (m Transaction) MarshalJSON() ([]byte, error) {
	raw := &rawTransaction{
		Input0:   m.Input0,
		Sig0:     hexOrNil(m.Sig0),
		Input1:   m.Input1,
		Sig1:     hexOrNil(m.Sig1),
		Output0:  m.Output0,
		Output1:  m.Output1,
		Fee:      m.Fee,
		BlockNum: m.BlockNum,
		TxIdx:    m.TxIdx,
		RootSig:  hexutil.Encode(m.RootSig),
	}
	return json.Marshal(raw)
}

type rawBlock struct {
	Header *BlockHeader `json:"header"`
	Hash   string       `json:"hash"`
}

func (m Block) MarshalJSON() ([]byte, error) {
	raw := &rawBlock{
		Header: m.Header,
		Hash:   hexutil.Encode(m.Hash),
	}
	return json.Marshal(raw)
}

type rawGetOutputsResponse struct {
	Transactions []*Transaction `json:"transactions"`
}

func (m GetOutputsResponse) MarshalJSON() ([]byte, error) {
	raw := &rawGetOutputsResponse{
		Transactions:m.Transactions,
	}
	return json.Marshal(raw)
}

func hexOrNil(b []byte) (*string) {
	if len(b) == 0 {
		return nil
	}

	out := hexutil.Encode(b)
	return &out
}

package chain

type BlockWithMeta struct {
	Block                 *Block                 `json:"block"`
	Metadata              *BlockMetadata         `json:"metadata"`
	ConfirmedTransactions []ConfirmedTransaction `json:"confirmedTransactions"`
}

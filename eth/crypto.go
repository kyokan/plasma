package eth

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
)

func GethHash(data []byte) []byte {
	msg := fmt.Sprintf("%s%d%s", SignaturePreamble, len(data), data)
	return crypto.Keccak256([]byte(msg))
}

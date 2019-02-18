package rpc

import (
	"encoding/hex"
	"github.com/kyokan/plasma/pkg/rpc/pb"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"strings"
	"fmt"
)

func SerializeBig(in *big.Int) (*pb.BigInt) {
	return &pb.BigInt{
		Hex: fmt.Sprintf("0x%s", strings.ToLower(in.Text(16))),
	}
}

func DeserializeBig(in *pb.BigInt) (*big.Int) {
	s := hex.EncodeToString(common.FromHex(in.Hex)) // Ox trips big.Int.SetString
	if len(s) == 0 {
		return big.NewInt(0)
	}
	b, _ := new(big.Int).SetString(s, 16)
	return b
}

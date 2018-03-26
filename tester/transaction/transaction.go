package transaction

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/chain"
)

func ExampleEncoder() {
	var t *chain.Transaction // t is nil pointer to PlasmaTransaction
	bytes, _ := rlp.EncodeToBytes(t)
	fmt.Printf("%v → %X\n", t, bytes)

	t = &chain.Transaction{
		Output0: &chain.Output{
			NewOwner: common.HexToAddress("1421e90e-1b4b-4f07-872e-20178c2c2b12"),
			Amount:   new(big.Int).SetUint64(5),
		},
		Output1: &chain.Output{
			NewOwner: common.HexToAddress("c81c342b-4fb0-46a0-9c6d-9688031e4854"),
			Amount:   new(big.Int).SetUint64(6),
		},
	}

	bytes, _ = rlp.EncodeToBytes(t)
	fmt.Printf("%v → %X\n", t, bytes)

	// Output:
	// <nil> → C28080
	// &{foobar 5 6} → C20506
	// &{<nil> <nil> [] [] 0xc42036e160 0xc42036e180 <nil> 0 0} → EC94000000000000000000000000000000001421E90E059400000000000000000000000000000000C81C342B06
}

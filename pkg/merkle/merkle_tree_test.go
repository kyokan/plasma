package merkle

import (
	"testing"
		"github.com/kyokan/plasma/util"
		"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
)

func TestRootAndProof3LeavesIdx0(t *testing.T) {
	leaf1 := util.Keccak256([]byte("inputSeed1"))
	leaf2 := util.Keccak256([]byte("inputSeed2"))
	leaf3 := util.Keccak256([]byte("inputSeed3"))

	root, proof := RootAndProof([]util.Hash{
		leaf1,
		leaf2,
		leaf3,
	}, 0)

	require.Equal(t, hexutil.Encode(root), "0xf323e1445d86cafa60e816c46696f2bd260f795fd76325b9c4aea430b668416a")
	require.Equal(t, hexutil.Encode(proof), "0xda2a5c2c3f35f3257026c91ae38f612664abd5c8a59135febb7ce80609013a261d7a0fdf5b4e222d2b55fcdb7f24bbda511e961bd71e54ca8204bb33a7cbd9d9")
}

func TestRootAndProof5LeavesIdx3(t *testing.T) {
	leaf1 := util.Keccak256([]byte("inputSeed1"))
	leaf2 := util.Keccak256([]byte("inputSeed2"))
	leaf3 := util.Keccak256([]byte("inputSeed3"))
	leaf4 := util.Keccak256([]byte("inputSeed4"))
	leaf5 := util.Keccak256([]byte("inputSeed5"))

	root, proof := RootAndProof([]util.Hash{
		leaf1,
		leaf2,
		leaf3,
		leaf4,
		leaf5,
	}, 3)

	require.Equal(t, hexutil.Encode(root), "0xffbaed089612123586b4f74e04cb0a02e300f33551182eaff703f4feda8927f4")
	require.Equal(t, hexutil.Encode(proof), "0xb3ca086bdc6d4011c11b58e1790c6b50800069a25eeace6ba73c580694e7e2c6f323e1445d86cafa60e816c46696f2bd260f795fd76325b9c4aea430b668416a")
}

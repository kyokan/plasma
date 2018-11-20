package merkle

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/util"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

type TestHashable struct {
	counter uint8
}

func doHash(i uint8) util.Hash {
	var v uint64
	v = 1 << i
	hash := fmt.Sprintf("%032x", v)
	return common.Hex2Bytes(hash)
}

func (h TestHashable) Hash() util.Hash {
	return doHash(h.counter)
}

func (h TestHashable) RLPHash() util.Hash {
	return doHash(h.counter)
}

func (h TestHashable) SetIndex(index uint32) {
	h.counter = uint8(index)
}

func testHasher(input []byte) util.Hash {
	half := len(input) / 2
	output := make([]byte, half)
	for i := 0; i < half; i++ {
		output[i] = input[i] | input[half + i]
	}
	return output
}

func Test_NoTree(t *testing.T) {
	_, err := createMerkleQueue(testHasher, 1, -1, false)
	require.Error(t, err)
}

func Test_TreeTooDeep(t *testing.T) {
	_, err := createMerkleQueue(testHasher, maxDepth + 1, -1, false)
	require.Error(t, err)
}

func Test_FullTree(t *testing.T) {
	queue, err := createMerkleQueue(testHasher, 3, -1, false)
	require.NoError(t, err)
	for i := uint8(0); i < 4; i++ {
		h := TestHashable{counter: i}
		err = queue.Enqueue(h)
		require.NoError(t, err)
	}
	_, err = queue.GetRootHash()
	require.NoError(t, err)
}

func Test_Concurrency(t *testing.T) {
	udepth := uint32(7)
	queue, err := createMerkleQueue(testHasher, int32(udepth), -1, false)
	require.NoError(t, err)

	var wg sync.WaitGroup

	concurrency := uint8(64)

	wg.Add(int(concurrency))
	elements := make([]TestHashable, concurrency)
	for i := uint8(0); i < concurrency; i++ {
		elements[i].counter = i
	}
	fn := func(element TestHashable) {
		queue.Enqueue(element)
		wg.Done()
	}
	for i := uint8(0); i < concurrency; i++ {
		go fn(elements[i])
	}

	wg.Wait()
	root, err := queue.GetRootRLPHash()
	require.NoError(t, err)
	// This ensures that the indices are assigned sequentially.
	// If they weren't the root (bitwise) hash would be different.
	require.EqualValues(t, "0000000000000000ffffffffffffffff", common.Bytes2Hex(root))
}

func Test_PartialTree(t *testing.T) {
	depth := int32(maxDepth)

	queue, err := createMerkleQueue(testHasher, depth, -1, false)
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		var hash util.Hash

		for count := uint8(0); count < 5; count++ {
			h := TestHashable{counter: count}
			err = queue.Enqueue(h)
			require.NoError(t, err)
		}

		hash, err = queue.GetRootHash()
		require.NoError(t, err)
		require.NotEmpty(t, hash)
		require.EqualValues(t, uint8(31), hash[15])

		hash, err = queue.GetRootRLPHash()
		require.NoError(t, err)
		require.NotEmpty(t, hash)
		require.EqualValues(t, uint8(31), hash[15])

		queue.Reset()
	}
}

func Test_Proof(t *testing.T) {
	depth := int32(4)
	queue, err := createMerkleQueue(testHasher, depth, 2, false)
	require.NoError(t, err)

	var proof util.Hash
	for i := 0; i < 10; i++ {
		for count := uint8(0); count < 5; count++ {
			h := TestHashable{counter: count}
			err = queue.Enqueue(h)
			require.NoError(t, err)
		}

		proof, err = queue.GetRLPProof()
		require.NoError(t, err)
		require.EqualValues(t, 16 * (depth - 1), len(proof))
		require.EqualValues(t,  uint8(8), proof[0 * 16 + 15])
		require.EqualValues(t,  uint8(3), proof[1 * 16 + 15])
		require.EqualValues(t, uint8(16), proof[2 * 16 + 15])
		queue.Reset()
	}
}

func Test_DoGetProof(t *testing.T) {
	depth := int32(4)
	size  := uint8(5)
	input := make([]DualHashable, size)
	for i := uint8(0); i < size; i++ {
		input[i] = TestHashable{ counter: i}
	}
	proof, err := doGetProof(input, testHasher, depth, 2)
	require.NoError(t, err)
	require.EqualValues(t, 16 * (depth - 1), len(proof))
	require.EqualValues(t,  uint8(8), proof[0 * 16 + 15])
	require.EqualValues(t,  uint8(3), proof[1 * 16 + 15])
	require.EqualValues(t, uint8(16), proof[2 * 16 + 15])
}

func Test_OverCapacityTree(t *testing.T) {
	queue, err := createMerkleQueue(testHasher, 3, -1, false)
	require.NoError(t, err)
	for i := uint8(0); i < 4; i++ {
		h := TestHashable{counter: i}
		err = queue.Enqueue(h)
		require.NoError(t, err)
	}
	h := TestHashable{counter: 4}
	err = queue.Enqueue(h)
	require.Error(t, err)
	_, err = queue.GetRootHash()
	require.NoError(t, err)
}

func Test_EmptyTree(t *testing.T) {
	queue, err := NewMerkleQueue(3, false)
	require.NoError(t, err)
	_, err = queue.GetRootHash()
	require.NoError(t, err)
}

func Test_PrecomputeHashes(t *testing.T) {
	expectedValues := [maxDepth]string{
		"776a31db34a1a0a7caaf862cffdfff1789297ffadc380bd3d39281d340abd3ad",
		"2def10d13dd169f550f578bda343d9717a138562e0093b380a1120789d53cf10",
		"4ebfd9cd7bca2505f7bef59cc1c12ecc708fff26ae4af19abe852afe9e20c862",
		"a2fca4a49658f9fab7aa63289c91b7c7b6c832a6d0e69334ff5b0a3483d09dab",
		"ad676aa337a485e4728a0b240d92b3ef7b3c372d06d189322bfd5f61f1e7203e",
		"3d04cffd8b46a874edf5cfae63077de85f849a660426697b06a829c70dd1409c",
		"e026cc5a4aed3c22a58cbd3d2ac754c9352c5436f638042dca99034e83636516",
		"7ad66c0a68c72cb89e4fb4303841966e4062a76ab97451e3b9fb526a5ceb7f82",
		"e1cea92ed99acdcb045a6726b2f87107e8a61620a232cf4d7d5b5766b3952e10",
		"292c23a9aa1d8bea7e2435e555a4a60e379a5a35f3f452bae60121073fb6eead",
		"617bdd11f7c0a11f49db22f629387a12da7596f9d1704d7465177c63d88ec7d7",
		"defff6d330bb5403f63b14f33b578274160de3a50df4efecf0e0db73bcdd3da5",
		"ecd50eee38e386bd62be9bedb990706951b65fe053bd9d8a521af753d139e2da",
		"3b8ec09e026fdc305365dfc94e189a81b38c7597b3d941c279f042e8206e0bd8",
		"890740a8eb06ce9be422cb8da5cdafc2b58c0a5e24036c578de2a433c828ff7d",
		"633dc4d7da7256660a892f8f1604a44b5432649cc8ec5cb3ced4c4e6ac94dd1d",
		"290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563",
	}
	precomputeHashes(util.DoHash, false)
	for i := 0; i < maxDepth; i++ {
		require.EqualValues(t, expectedValues[i], common.Bytes2Hex(precomputedHashes[i]))
	}
	precomputeHashes(util.DoHash, true)
	for i := 0; i < maxDepth; i++ {
		require.EqualValues(t, expectedValues[maxDepth - 1], common.Bytes2Hex(precomputedHashes[i]))
	}
	zero := common.Bytes2Hex(make([]byte, 16))
	precomputeHashes(testHasher, true)
	for i := 0; i < maxDepth; i++ {
		require.EqualValues(t, zero, common.Bytes2Hex(precomputedHashes[i]))
	}
	precomputeHashes(testHasher, false)
	for i := 0; i < maxDepth; i++ {
		require.EqualValues(t, zero, common.Bytes2Hex(precomputedHashes[i]))
	}
}


type Tx struct {
	rlp string
}

func (t Tx) Hash() util.Hash {
	return util.DoHash(common.Hex2Bytes(t.rlp))
}

func (t Tx) RLPHash() util.Hash {
	return util.DoHash(common.Hex2Bytes(t.rlp))
}

func (t Tx) SetIndex(uint32) {

}

func Test_ShallowTree(t *testing.T) {
	tx := &Tx{
		rlp: "f90106028080b8402e44fe9dbe39750a4c62bb365dcdb658c5b99ce529664ed99a4ad47afc6982f544efa08b8e808aa4ba7f3506af9ca2cd09c54b4f88fd965431fdc526c4bba37b808080b8402e44fe9dbe39750a4c62bb365dcdb658c5b99ce529664ed99a4ad47afc6982f544efa08b8e808aa4ba7f3506af9ca2cd09c54b4f88fd965431fdc526c4bba37b94f17f52151ebef6c7334fad080c5704d77216b732843b9aca0094627306090abab3a6e1400e9345bc60c78a8bef578829a2241aba91360080b8414854aff1aa2e81bc2355ea1f73ab7eb374f93278dd7e7ba901c777b48c13605c14ac59ef14d1507ff0bed63483c43794c483ef388f2d15ca4fa8287dd74f21f601",
	}
	precomputeHashes(util.DoHash, false)
	queue, err := createMerkleQueue(util.DoHash, 3, -1, false)
	require.NoError(t, err)
	queue.Enqueue(tx)
	root, err := queue.GetRootRLPHash()
	require.NoError(t, err)
	precomputeHashes(testHasher, false)
	require.EqualValues(t, "9a129393f9509d32aaa85f251959627582bb42cbd498794c6529bba39b9bb38b", common.Bytes2Hex(root))
}

func Test_DoHash(t *testing.T) {
	output := util.DoHash([]byte{})
	// If fails the underlying hash function has changed
	// SHA-3 output for empty input is here https://csrc.nist.gov/csrc/media/projects/cryptographic-standards-and-guidelines/documents/examples/sha3-256_msg0.pdf
	// The value expected here is for the original Keccak algorithm
	// See more here https://medium.com/@ConsenSys/are-you-really-using-sha-3-or-old-code-c5df31ad2b0
	require.EqualValues(t, "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470", common.Bytes2Hex(output))
}
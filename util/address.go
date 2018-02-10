package util

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

func AddressesEqual(addr1 *common.Address, addr2 *common.Address) bool {
	return bytes.Equal(addr1.Bytes(), addr2.Bytes())
}

func AddressToHex(addr *common.Address) string {
	return strings.ToLower(addr.Hex())
}

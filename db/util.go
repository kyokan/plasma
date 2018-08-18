package db

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

)

const keyPartsSeparator = "::"

func prefixKey(prefix string, parts ...string) []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte(prefix))

	for _, part := range parts {
		buf.Write([]byte(keyPartsSeparator))
		buf.Write([]byte(part))
	}

	return buf.Bytes()
}

func uint64ToBytes(i uint64) []byte {
	return []byte(fmt.Sprintf("%X", i))
}

func bytesToUint64(b []byte) uint64 {
	s := string(b)

	n, err := strconv.ParseUint(s, 16, 32)

	if err != nil {
		log.Fatalf("Failed to parse string as hex: %v", err)
	}

	return uint64(n)
}

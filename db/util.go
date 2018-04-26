package db

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type GuardedDb struct {
	db  *leveldb.DB
	err error
}

func (gd *GuardedDb) Put(key []byte, value []byte, wo *opt.WriteOptions) {
	if gd.err != nil {
		return
	}

	gd.err = gd.db.Put(key, value, wo)
}

func (gd *GuardedDb) Get(key []byte, ro *opt.ReadOptions) []byte {
	if gd.err != nil {
		return nil
	}

	data, err := gd.db.Get(key, ro)
	gd.err = err
	return data
}

func (gd *GuardedDb) Has(key []byte, ro *opt.ReadOptions) bool {
	if gd.err != nil {
		return false
	}

	data, err := gd.db.Has(key, ro)
	gd.err = err
	return data
}

func prefixKey(prefix string, parts ...string) []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte(prefix))

	for _, part := range parts {
		buf.Write([]byte("::"))
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

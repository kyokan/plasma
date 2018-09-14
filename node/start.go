package node

import (
	"encoding/hex"
	"log"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"gopkg.in/urfave/cli.v1"
)

func ForceSubmitBlock(c *cli.Context) {
	merkleRoot, err := hex.DecodeString(c.String("merkle-root"))

	if err != nil {
		log.Fatal(err)
	}

	prevHash, err := hex.DecodeString(c.String("prev-hash"))

	if err != nil {
		log.Fatal(err)
	}

	number := c.Int("number")

	dburl := c.GlobalString("db")

	db, storage, err := db.CreateStorage(dburl, nil)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	log.Println("Received ForceSubmitBlock request.")

	header := chain.BlockHeader{
		MerkleRoot:    merkleRoot,
		RLPMerkleRoot: merkleRoot,
		PrevHash:      prevHash,
		Number:        uint64(number),
	}

	block := chain.Block{
		Header:    &header,
		BlockHash: header.Hash(),
	}

	if err := storage.SaveBlock(&block); err != nil {
		log.Fatalf("Failed to create genesis block:%v\n", err)
	}

	log.Printf("Submitted block with hash: %s\n", hex.EncodeToString(block.BlockHash))
}

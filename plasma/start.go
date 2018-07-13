package plasma

import (
	"encoding/hex"
	"log"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/node"
	"github.com/kyokan/plasma/rpc"
	"gopkg.in/urfave/cli.v1"
)

func Start(c *cli.Context) {
	nodeURL := c.GlobalString("node-url")
	dburl := c.GlobalString("db")

	plasma := eth.CreatePlasmaClientCLI(c)

	db, storage, err := db.CreateStorage(dburl, plasma)

	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	client, err := eth.NewClient(nodeURL)
	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	sink := node.NewTransactionSink(storage, client)

	p := node.NewPlasmaNode(storage, sink, plasma)

	go p.Start()

	go rpc.Start(c.Int("rpc-port"), storage, sink)

	// TODO: ensure that 1 deposit tx is always 1 block
	go node.StartDepositListener(storage, sink, plasma)

	// TODO: add an exit listener to make sure to add an exit transaction to root node.
	// Also add an exit block to the plasma contract.

	select {}
}

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

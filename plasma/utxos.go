package plasma

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/db"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"log"
	"os"
)

func PrintUTXOs(c *cli.Context) {
	level, err := db.CreateLevelDatabase(c.GlobalString("db"))

	if err != nil {
		log.Panic("Failed to establish connection with database:", err)
	}

	addrStr := c.String("addr")

	if addrStr == "" {
		log.Panic("Addr is required.")
	}

	addr := common.HexToAddress(c.String("addr"))
	txs, err := level.AddressDao.UTXOs(&addr)

	if err != nil {
		log.Panic("Failed to get UTXOs: ", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Hash", "Amount", "Block Number", "Tx Index"})
	for _, tx := range txs {
		table.Append([]string{
			common.ToHex(tx.Hash()),
			tx.OutputFor(&addr).Amount.String(),
			fmt.Sprint(tx.BlkNum),
			fmt.Sprint(tx.TxIdx),
		})
	}

	table.Render()
}

package plasma

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/db"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

// TODO: migrate to root userclient.
func PrintUTXOs(c *cli.Context) {
	db, level, err := db.CreateLevelDatabase(c.GlobalString("db"))

	if err != nil {
		log.Panic("Failed to establish connection with database:", err)
	}

	defer db.Close()

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

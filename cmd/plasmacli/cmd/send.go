package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ethereum/go-ethereum/crypto"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/kyokan/plasma/log"
	"crypto/ecdsa"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/eth"
	"time"
	"github.com/kyokan/plasma/rpc/pb"
	"context"
	"bytes"
	"github.com/kyokan/plasma/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"fmt"
	"sort"
)

type sendCmdOutput struct {
	Value            string   `json:"value"`
	To               string   `json:"to"`
	BlockNumber      uint64   `json:"blockNumber"`
	TransactionIndex uint32   `json:"transactionIndex"`
	DepositNonce     string   `json:"depositNonce"`
	MerkleRoot       string   `json:"merkleRoot"`
	ConfirmSigs      []string `json:"confirmSigs"`
}

var sendCmdLog = log.ForSubsystem("SendCmd")

var sendCmd = &cobra.Command{
	Use:   "send to value [depositNonce] [contractAddr]",
	Short: "Sends funds",
	RunE: func(cmd *cobra.Command, args []string) error {
		privKey, err := ParsePrivateKey(cmd)
		if err != nil {
			return err
		}
		from := crypto.PubkeyToAddress(privKey.PublicKey)
		to := common.HexToAddress(args[0])
		value, ok := new(big.Int).SetString(args[1], 10)
		if !ok {
			return errors.New("invalid send value")
		}

		if len(args) == 4 {
			depositNonce, ok := new(big.Int).SetString(args[2], 10)
			if !ok {
				return errors.New("invalid deposit nonce")
			}
			contractAddr := common.HexToAddress(args[3])
			return spendDeposit(cmd, privKey, from, to, value, depositNonce, contractAddr)
		}

		return spendTx(cmd, privKey, from, to, value)
	},
}

func spendDeposit(cmd *cobra.Command, privKey *ecdsa.PrivateKey, from common.Address, to common.Address, value *big.Int, depositNonce *big.Int, contractAddr common.Address) error {
	client, conn, err := CreateRootClient(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()

	sendCmdLog.Info("spending deposit")
	contract, err := eth.NewClient(cmd.Flag(FlagEthereumNodeUrl).Value.String(), contractAddr.Hex(), privKey)
	if err != nil {
		return err
	}
	total, owner, err := contract.LookupDeposit(depositNonce)
	if err != nil {
		return err
	}
	if owner != from {
		return errors.New("you don't own this deposit")
	}
	if total.Cmp(value) < 0 {
		return errors.New("cannot send more than the deposit amount")
	}

	body := chain.ZeroBody()
	body.Input0.DepositNonce = depositNonce
	body.Output0.Amount = value
	body.Output0.Owner = to
	if total.Cmp(body.Output0.Amount) > 0 {
		totalClone := new(big.Int).Set(total)
		body.Output1.Amount = totalClone.Sub(totalClone, value)
		body.Output1.Owner = from
	}

	fmt.Println("sighash", hexutil.Encode(body.SignatureHash()))
	sig, err := eth.Sign(privKey, body.SignatureHash())
	if err != nil {
		return err
	}

	// no confirm sigs on deposits
	tx := &chain.Transaction{
		Body: body,
		Sigs: [2]chain.Signature{
			sig,
			sig,
		},
	}

	sendCmdLog.Info("sending spend message")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	sendRes, err := client.Send(ctx, &pb.SendRequest{
		Transaction: tx.Proto(),
	})
	if err != nil {
		return err
	}

	sendCmdLog.Info("confirming spend")

	tx.Body.BlockNumber = sendRes.Inclusion.BlockNumber
	tx.Body.TransactionIndex = sendRes.Inclusion.TransactionIndex
	var buf bytes.Buffer
	buf.Write(tx.RLPHash(util.Sha256))
	buf.Write(sendRes.Inclusion.MerkleRoot)
	sigHash := util.Sha256(buf.Bytes())
	confirmSig, err := eth.Sign(privKey, sigHash)
	if err != nil {
		return err
	}

	ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
	_, err = client.Confirm(ctx, &pb.ConfirmRequest{
		BlockNumber:      sendRes.Inclusion.BlockNumber,
		TransactionIndex: sendRes.Inclusion.TransactionIndex,
		ConfirmSig0:      confirmSig[:],
		ConfirmSig1:      confirmSig[:],
	})
	if err != nil {
		return err
	}

	out := &sendCmdOutput{
		Value:            value.Text(10),
		To:               to.Hex(),
		BlockNumber:      sendRes.Inclusion.BlockNumber,
		TransactionIndex: sendRes.Inclusion.TransactionIndex,
		MerkleRoot:       hexutil.Encode(sendRes.Inclusion.MerkleRoot),
		ConfirmSigs: []string{
			hexutil.Encode(confirmSig[:]),
			hexutil.Encode(confirmSig[:]),
		},
	}

	return PrintJSON(out)
}

func spendTx(cmd *cobra.Command, privKey *ecdsa.PrivateKey, from common.Address, to common.Address, value *big.Int) error {
	client, conn, err := CreateRootClient(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()

	sendCmdLog.Info("selecting outputs")

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	res, err := client.GetOutputs(ctx, &pb.GetOutputsRequest{
		Address:   from.Bytes(),
		Spendable: true,
	})
	if err != nil {
		return err
	}

	var utxos []chain.ConfirmedTransaction
	for _, utxoProto := range res.ConfirmedTransactions {
		confirmedTx, err := chain.ConfirmedTransactionFromProto(utxoProto)
		if err != nil {
			return err
		}
		utxos = append(utxos, *confirmedTx)
	}
	if len(utxos) == 0 {
		return errors.New("no spendable outputs")
	}
	selectedUTXOs, err := selectUTXOs(utxos, from, value)
	if err != nil {
		return err
	}

	total := big.NewInt(0)
	tx := &chain.Transaction{
		Body: chain.ZeroBody(),
	}
	for i, utxo := range selectedUTXOs {
		txBody := utxo.Transaction.Body
		var input *chain.Input
		if i == 0 {
			input = tx.Body.Input0
		} else if i == 1 {
			input = tx.Body.Input1
		} else {
			panic("too many inputs!")
		}

		input.BlockNum = txBody.BlockNumber
		input.TxIdx = txBody.TransactionIndex
		input.OutIdx = txBody.OutputIndexFor(&from)

		if i == 0 {
			tx.Body.Input0ConfirmSig = utxo.ConfirmSigs[0]
		} else {
			tx.Body.Input1ConfirmSig = utxo.ConfirmSigs[1]
		}

		total = total.Add(total, txBody.OutputFor(&from).Amount)
	}

	tx.Body.Output0.Amount = value
	tx.Body.Output0.Owner = to

	if total.Cmp(value) > 0 {
		totalClone := new(big.Int).Set(total)
		tx.Body.Output1.Amount = totalClone.Sub(totalClone, value)
		tx.Body.Output1.Owner = from
	}

	sig, err := eth.Sign(privKey, tx.Body.SignatureHash())
	if err != nil {
		return err
	}
	tx.Sigs[0] = sig
	tx.Sigs[1] = sig

	ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
	sendRes, err := client.Send(ctx, &pb.SendRequest{
		Transaction: tx.Proto(),
	})
	if err != nil {
		return err
	}

	tx.Body.BlockNumber = sendRes.Inclusion.BlockNumber
	tx.Body.TransactionIndex = sendRes.Inclusion.TransactionIndex
	var buf bytes.Buffer
	buf.Write(tx.RLPHash(util.Sha256))
	buf.Write(sendRes.Inclusion.MerkleRoot)
	sigHash := util.Sha256(buf.Bytes())
	confirmSig, err := eth.Sign(privKey, sigHash)
	if err != nil {
		return err
	}
	fmt.Println(hexutil.Encode(buf.Bytes()))

	sendCmdLog.Info("confirming transaction")

	ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
	_, err = client.Confirm(ctx, &pb.ConfirmRequest{
		BlockNumber:      sendRes.Inclusion.BlockNumber,
		TransactionIndex: sendRes.Inclusion.TransactionIndex,
		ConfirmSig0:      confirmSig[:],
		ConfirmSig1:      confirmSig[:],
	})
	if err != nil {
		return err
	}

	out := &sendCmdOutput{
		Value:            value.Text(10),
		To:               to.Hex(),
		BlockNumber:      sendRes.Inclusion.BlockNumber,
		TransactionIndex: sendRes.Inclusion.TransactionIndex,
		MerkleRoot:       hexutil.Encode(sendRes.Inclusion.MerkleRoot),
		ConfirmSigs: []string{
			hexutil.Encode(confirmSig[:]),
			hexutil.Encode(confirmSig[:]),
		},
	}

	return PrintJSON(out)
}

func selectUTXOs(confirmedTxs []chain.ConfirmedTransaction, addr common.Address, total *big.Int) ([]chain.ConfirmedTransaction, error) {
	sort.Slice(confirmedTxs, func(i, j int) bool {
		a := confirmedTxs[i].Transaction.Body.OutputFor(&addr).Amount
		b := confirmedTxs[j].Transaction.Body.OutputFor(&addr).Amount
		return a.Cmp(b) > 0
	})

	first := confirmedTxs[0]
	firstBody := first.Transaction.Body

	if firstBody.OutputFor(&addr).Amount.Cmp(total) >= 0 {
		return []chain.ConfirmedTransaction{first}, nil
	}

	for i := len(confirmedTxs) - 1; i >= 0; i-- {
		second := confirmedTxs[i]
		secondBody := second.Transaction.Body
		sum := big.NewInt(0)
		sum = sum.Add(sum, firstBody.OutputFor(&addr).Amount)
		sum = sum.Add(sum, secondBody.OutputFor(&addr).Amount)
		if sum.Cmp(total) >= 0 {
			return []chain.ConfirmedTransaction{
				first,
				second,
			}, nil
		}
	}

	return nil, errors.New("no suitable UTXOs found")
}

func init() {
	rootCmd.AddCommand(sendCmd)
	sendCmd.Flags().StringP(FlagEthereumNodeUrl, "e", "http://localhost:8545", "URL to a running Ethereum node.")
}

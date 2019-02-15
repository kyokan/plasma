package root

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/rpc"
	"github.com/kyokan/plasma/rpc/pb"
	"google.golang.org/grpc"
	"log"
	"net"
	"github.com/kyokan/plasma/node"
	"github.com/pkg/errors"
	log2 "github.com/kyokan/plasma/log"
)

type Server struct {
	storage   db.PlasmaStorage
	ctx       context.Context
	mPool     *node.Mempool
	confirmer *node.TransactionConfirmer
}

var logger = log2.ForSubsystem("RootServer")

func NewServer(ctx context.Context, storage db.PlasmaStorage, mPool *node.Mempool, confirmer *node.TransactionConfirmer) (*Server) {
	return &Server{
		storage:   storage,
		ctx:       ctx,
		mPool:     mPool,
		confirmer: confirmer,
	}
}

func (r *Server) Start(rpcPort int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		log.Println("error", err)
		return err
	}

	s := grpc.NewServer()
	pb.RegisterRootServer(s, r)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Println("error", err)
		}
	}()

	go func() {
		<-r.ctx.Done()
		s.Stop()
	}()

	log.Printf("Started RPC server on port %d", rpcPort)

	return nil
}

func (r *Server) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	addr := common.BytesToAddress(req.Address)
	bal, err := r.storage.Balance(addr)
	if err != nil {
		return nil, err
	}

	return &pb.GetBalanceResponse{
		Balance: rpc.SerializeBig(bal),
	}, nil
}

func (r *Server) GetOutputs(ctx context.Context, req *pb.GetOutputsRequest) (*pb.GetOutputsResponse, error) {
	addr := common.BytesToAddress(req.Address)
	var txs []chain.ConfirmedTransaction
	var err error
	if req.Spendable {
		txs, err = r.storage.SpendableTxs(addr)
	} else {
		txs, err = r.storage.UTXOs(addr)
	}

	if err != nil {
		return nil, err
	}

	var ret []*pb.ConfirmedTransaction
	for _, tx := range txs {
		ret = append(ret, tx.Proto())
	}

	return &pb.GetOutputsResponse{
		ConfirmedTransactions: ret,
	}, nil
}

func (r *Server) GetBlock(ctx context.Context, req *pb.GetBlockRequest) (*pb.GetBlockResponse, error) {
	block, err := r.storage.BlockAtHeight(req.Number)
	if err != nil {
		log2.WithError(logger, err).Error("failed to fetch block at height")
		return nil, err
	}
	txs, err := r.storage.FindTransactionsByBlockNum(block.Header.Number)
	if err != nil {
		log2.WithError(logger, err).Error("failed to fetch transactions")
		return nil, err
	}
	meta, err := r.storage.BlockMetaAtHeight(req.Number)
	if err != nil {
		return nil, err
	}
	var confirmedTxs []*pb.ConfirmedTransaction
	for _, tx := range txs {
		confirmedTxs = append(confirmedTxs, tx.Proto())
	}

	res := &pb.GetBlockResponse{
		Block: &pb.Block{
			Header: &pb.BlockHeader{
				MerkleRoot:    block.Header.MerkleRoot,
				RlpMerkleRoot: block.Header.RLPMerkleRoot,
				PrevHash:      block.Header.PrevHash,
				Number:        req.Number,
			},
			Hash: block.BlockHash,
		},
		ConfirmedTransactions: confirmedTxs,
		Metadata: &pb.GetBlockResponse_BlockMeta{
			CreatedAt: meta.CreatedAt,
		},
	}

	return res, nil
}

func (r *Server) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	if req == nil {
		return nil, errors.New("no request provided")
	}

	tx, err := chain.TransactionFromProto(req.Transaction)
	if err != nil {
		return nil, err
	}

	inclusion := r.mPool.Append(*tx)
	if inclusion.Error != nil {
		return nil, inclusion.Error
	}

	tx.Body.BlockNumber = inclusion.BlockNumber
	tx.Body.TransactionIndex = inclusion.TransactionIndex

	return &pb.SendResponse{
		Transaction: tx.Proto(),
		Inclusion: &pb.TransactionInclusion{
			MerkleRoot:       inclusion.MerkleRoot[:],
			BlockNumber:      inclusion.BlockNumber,
			TransactionIndex: inclusion.TransactionIndex,
		},
	}, nil
}

func (r *Server) Confirm(ctx context.Context, req *pb.ConfirmRequest) (*pb.ConfirmedTransaction, error) {
	var sig0 chain.Signature
	copy(sig0[:], req.ConfirmSig0)
	var sig1 chain.Signature
	copy(sig1[:], req.ConfirmSig1)

	tx, err := r.confirmer.Confirm(req.BlockNumber, req.TransactionIndex, [2]chain.Signature{
		sig0,
		sig1,
	})
	if err != nil {
		return nil, err
	}

	return tx.Proto(), nil
}

func (r *Server) GetConfirmations(ctx context.Context, req *pb.GetConfirmationsRequest) (*pb.GetConfirmationsResponse, error) {
	sigs, err := r.confirmer.GetConfirmations(req.Sig, req.Nonce, req.BlockNumber, req.TransactionIndex, uint8(req.OutputIndex))
	if err != nil {
		return nil, err
	}

	return &pb.GetConfirmationsResponse{
		AuthSig0: sigs[0][:],
		AuthSig1: sigs[1][:],
	}, nil
}

func (r *Server) BlockHeight(context.Context, *pb.EmptyRequest) (*pb.BlockHeightResponse, error) {
	latest, err := r.storage.LatestBlock()
	if err != nil {
		return nil, err
	}

	return &pb.BlockHeightResponse{
		Height: latest.Header.Number,
	}, nil
}

package root

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/pkg/chain"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/kyokan/plasma/pkg/rpc"
	"github.com/kyokan/plasma/pkg/rpc/pb"
	"google.golang.org/grpc"
	"net"
	"github.com/kyokan/plasma/pkg/service"
	"github.com/pkg/errors"
	"github.com/kyokan/plasma/pkg/log"
	"github.com/sirupsen/logrus"
)

type Server struct {
	storage   db.Storage
	ctx       context.Context
	mPool     *service.Mempool
	confirmer *service.TransactionConfirmer
}

var logger = log.ForSubsystem("RootServer")

func NewServer(ctx context.Context, storage db.Storage, mPool *service.Mempool, confirmer *service.TransactionConfirmer) (*Server) {
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
		return err
	}

	s := grpc.NewServer()
	pb.RegisterRootServer(s, r)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.WithError(logger, err)
		}
	}()

	go func() {
		<-r.ctx.Done()
		s.Stop()
	}()

	logger.WithFields(logrus.Fields{
		"port": rpcPort,
	}).Info("started gRPC server")

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
	return r.getBlock(req.Number)
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

func (r *Server) BlockHeight(context.Context, *pb.EmptyRequest) (*pb.BlockHeightResponse, error) {
	latest, err := r.storage.LatestBlock()
	if err != nil {
		return nil, err
	}

	return &pb.BlockHeightResponse{
		Height: latest.Header.Number,
	}, nil
}

func (r *Server) Sync(req *pb.SyncRequest, stream pb.Root_SyncServer) error {
	head, err := r.storage.LatestBlock()
	if err != nil {
		return err
	}
	if req.Start > head.Header.Number || req.Start == 0 {
		return nil
	}

	num := req.Start
	for num <= head.Header.Number {
		res, err := r.getBlock(num)
		if err != nil {
			return err
		}
		if err := stream.Send(res); err != nil {
			return err
		}
		num++
	}
	return nil
}

func (r *Server) getBlock(height uint64) (*pb.GetBlockResponse, error) {
	block, err := r.storage.BlockAtHeight(height)
	if err != nil {
		log.WithError(logger, err).Error("failed to fetch block at height")
		return nil, err
	}
	txs, err := r.storage.FindTransactionsByBlockNum(block.Header.Number)
	if err != nil {
		log.WithError(logger, err).Error("failed to fetch transactions")
		return nil, err
	}
	meta, err := r.storage.BlockMetaAtHeight(height)
	if err != nil {
		return nil, err
	}
	var confirmedTxs []*pb.ConfirmedTransaction
	for _, tx := range txs {
		confirmedTxs = append(confirmedTxs, tx.Proto())
	}

	return &pb.GetBlockResponse{
		Block:                 block.Proto(),
		ConfirmedTransactions: confirmedTxs,
		Metadata:              meta.Proto(),
	}, nil
}
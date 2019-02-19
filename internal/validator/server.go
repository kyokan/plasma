package validator

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
	"github.com/pkg/errors"
	"github.com/kyokan/plasma/pkg/log"
	"github.com/sirupsen/logrus"
	"time"
	"github.com/kyokan/plasma/pkg/service"
)

type Server struct {
	storage     db.Storage
	ctx         context.Context
	rootClient  pb.RootClient
	mainBreaker service.CircuitBreaker
}

var logger = log.ForSubsystem("ValidatorServer")

func NewServer(ctx context.Context, storage db.Storage, rootClient pb.RootClient, mainBreaker service.CircuitBreaker) (*Server) {
	return &Server{
		storage:     storage,
		ctx:         ctx,
		rootClient:  rootClient,
		mainBreaker: mainBreaker,
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
	if r.mainBreaker.Tripped() {
		return nil, errors.New("transactions are disable due to circuit breaker")
	}

	if req == nil {
		return nil, errors.New("no request provided")
	}

	childCtx, _ := context.WithTimeout(ctx, 5*time.Second)
	return r.rootClient.Send(childCtx, req)
}

func (r *Server) Confirm(ctx context.Context, req *pb.ConfirmRequest) (*pb.ConfirmedTransaction, error) {
	if r.mainBreaker.Tripped() {
		return nil, errors.New("confirmations are disable due to circuit breaker")
	}

	childCtx, _ := context.WithTimeout(ctx, 5*time.Second)
	return r.rootClient.Confirm(childCtx, req)
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

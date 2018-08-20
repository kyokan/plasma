package root

import (
	"context"
	"github.com/kyokan/plasma/rpc/pb"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/rpc"
	"github.com/ethereum/go-ethereum/common"
	"net"
	"fmt"
	"google.golang.org/grpc"
	"github.com/kyokan/plasma/node"
	"github.com/kyokan/plasma/types"
	"log"
)

type Server struct {
	storage db.PlasmaStorage
	ctx     context.Context
	sink    *node.TransactionSink
}

func NewServer(ctx context.Context, storage db.PlasmaStorage, sink *node.TransactionSink) (*Server) {
	return &Server{
		storage: storage,
		ctx:     ctx,
		sink:    sink,
	}
}

func (r *Server) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("error", err)
		return err
	}

	s := grpc.NewServer()
	pb.RegisterRootServer(s, r)
	if err := s.Serve(lis); err != nil {
		log.Println("error", err)
		return err
	}

	go func() {
		<-r.ctx.Done()
		s.Stop()
	}()

	return nil
}

func (r *Server) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	addr := common.BytesToAddress(req.Address)
	bal, err := r.storage.Balance(&addr)
	if err != nil {
		return nil, err
	}

	return &pb.GetBalanceResponse{
		Balance: rpc.SerializeBig(bal),
	}, nil
}

func (r *Server) GetUTXOs(ctx context.Context, req *pb.GetUTXOsRequest) (*pb.GetUTXOsResponse, error) {
	addr := common.BytesToAddress(req.Address)
	txs, err := r.storage.UTXOs(&addr)
	if err != nil {
		return nil, err
	}

	return &pb.GetUTXOsResponse{
		Transactions: rpc.SerializeTxs(txs),
	}, nil
}

func (r *Server) GetBlock(ctx context.Context, req *pb.GetBlockRequest) (*pb.GetBlockResponse, error) {
	num := rpc.DeserializeBig(req.Number)
	block, err := r.storage.BlockAtHeight(num.Uint64())
	if err != nil {
		return nil, err
	}

	txs, err := r.storage.FindTransactionsByBlockNum(block.Header.Number)
	if err != nil {
		return nil, err
	}

	res := &pb.GetBlockResponse{
		Block: &pb.Block{
			Header: &pb.BlockHeader{
				MerkleRoot:    block.Header.MerkleRoot,
				RlpMerkleRoot: block.Header.RLPMerkleRoot,
				PrevHash:      block.Header.PrevHash,
				Number:        rpc.SerializeBig(num),
			},
			Hash: block.BlockHash,
		},
		Transactions: rpc.SerializeTxs(txs),
	}

	return res, nil
}

func (r *Server) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	tx := rpc.DeserializeTx(req.Transaction)
	from := common.BytesToAddress(req.From)
	to := common.BytesToAddress(req.To)
	amount := rpc.DeserializeBig(req.Amount)

	txReq := &types.TransactionRequest{
		Transaction: *tx,
		From:        from,
		To:          to,
		Amount:      amount,
	}

	ch := make(chan types.TransactionRequest)
	txChan := make(chan chan types.TransactionRequest)
	r.sink.AcceptTransactionRequests(txChan)
	txChan <- ch
	ch <- *txReq
	res := <-ch
	close(ch)

	if res.Response.Error != nil {
		return nil, res.Response.Error
	}

	return &pb.SendResponse{
		Transaction: rpc.SerializeTx(res.Response.Transaction),
	}, nil
}

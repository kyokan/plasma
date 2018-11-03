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
	"github.com/gin-gonic/gin"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	storage db.PlasmaStorage
	ctx     context.Context
}

func NewServer(ctx context.Context, storage db.PlasmaStorage) (*Server) {
	return &Server{
		storage: storage,
		ctx:     ctx,
	}
}

func (r *Server) Start(rpcPort int, restPort int) error {
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
	if err = r.startREST(restPort); err != nil {
		return err
	}

	return nil
}

func (r *Server) startREST(port int) error {
	rtr := gin.Default()
	rtr.GET("/v1/balance/:address", func(c *gin.Context) {
		addr, err := hexutil.Decode(c.Param("address"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		res, err := r.GetBalance(c, &pb.GetBalanceRequest{
			Address: addr,
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, res)
	})
	rtr.GET("/v1/outputs/:address/:spendable", func(c *gin.Context) {
		addr, err := hexutil.Decode(c.Param("address"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		spendable := c.Param("spendable") == "true"
		res, err := r.GetOutputs(c, &pb.GetOutputsRequest{
			Address:   addr,
			Spendable: spendable,
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, res)
	})
	rtr.GET("/v1/block/:number", func(c *gin.Context) {
		numStr := c.Param("number")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		res, err := r.GetBlock(c, &pb.GetBlockRequest{
			Number: uint64(num),
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, res)
	})
	rtr.GET("/v1/blockheight", func(c *gin.Context) {
		res, err := r.BlockHeight(c, &pb.EmptyRequest{})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, res)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: rtr,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	go func() {
		<-r.ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}()

	log.Printf("started REST server on %d\n", port)

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

func (r *Server) GetOutputs(ctx context.Context, req *pb.GetOutputsRequest) (*pb.GetOutputsResponse, error) {
	addr := common.BytesToAddress(req.Address)
	var txs []chain.Transaction
	var err error
	if req.Spendable {
		txs, err = r.storage.SpendableTxs(&addr)
	} else {
		txs, err = r.storage.UTXOs(&addr)
	}

	if err != nil {
		return nil, err
	}

	return &pb.GetOutputsResponse{
		Transactions: rpc.SerializeTxs(txs),
	}, nil
}

func (r *Server) GetBlock(ctx context.Context, req *pb.GetBlockRequest) (*pb.GetBlockResponse, error) {
	block, err := r.storage.BlockAtHeight(req.Number)
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
				Number:        req.Number,
			},
			Hash: block.BlockHash,
		},
		Transactions: rpc.SerializeTxs(txs),
	}

	return res, nil
}

func (r *Server) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	tx := rpc.DeserializeTx(req.Transaction)

	signedTx, storeErr := r.storage.StoreTransaction(*tx)
	if storeErr != nil {
		return nil, storeErr
	}
	return &pb.SendResponse{
		Transaction: rpc.SerializeTx(signedTx),
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

package root

import (
	"github.com/gin-gonic/gin"
	"github.com/kyokan/plasma/pkg/log"
	"net/http"
	"context"
	"time"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/kyokan/plasma/util"
	"github.com/sirupsen/logrus"
	"fmt"
	"github.com/pkg/errors"
	"github.com/kyokan/plasma/pkg/chain"
	"github.com/kyokan/plasma/pkg/service"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/syndtr/goleveldb/leveldb"
)

var restLogger = log.ForSubsystem("RESTServer")

type RESTServer struct {
	storage   db.Storage
	mpool     *service.Mempool
	confirmer *service.TransactionConfirmer
	port      int

	server *http.Server
	engine *gin.Engine
}

type WrappableHandlerFunc func(c *gin.Context) (interface{}, error)

type ConfirmationRequest struct {
	BlockNumber      uint64 `json:"blockNumber"`
	TransactionIndex uint32 `json:"transactionIndex"`
	ConfirmSig0      string `json:"confirmSig0"`
	ConfirmSig1      string `json:"confirmSig1"`
}

func NewRESTServer(storage db.Storage, mpool *service.Mempool, confirmer *service.TransactionConfirmer, port int) *RESTServer {
	return &RESTServer{
		storage:   storage,
		mpool:     mpool,
		confirmer: confirmer,
		port:      port,
	}
}

func (r *RESTServer) Start() error {
	r.engine = gin.Default()
	r.engine.GET("/balances/:address", r.wrapHandler(r.GetBalance))
	r.engine.GET("/utxos/:address", r.wrapHandler(r.GetUTXOs))
	r.engine.GET("/blocks/:height", r.wrapHandler(r.GetBlock))
	r.engine.POST("/send", r.wrapHandler(r.Send))
	r.engine.POST("/confirm", r.wrapHandler(r.Confirm))
	r.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", r.port),
		Handler: r.engine,
	}

	go func() {
		if err := r.server.ListenAndServe(); err != nil {
			log.WithError(restLogger, err).Error("encountered error in rest server")
			return
		}
	}()

	restLogger.WithFields(logrus.Fields{
		"port": r.port,
	}).Info("started REST server")

	return nil
}

func (r *RESTServer) Stop() error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return r.server.Shutdown(ctx)
}

func (r *RESTServer) GetBalance(c *gin.Context) (interface{}, error) {
	addrStr, exists := c.Params.Get("address")
	if !exists {
		return nil, errors.New("address is required")
	}
	addr := common.HexToAddress(addrStr)
	bal, err := r.storage.Balance(addr)
	if err == leveldb.ErrNotFound {
		return &gin.H{
			"balance": "0",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &gin.H{
		"balance": util.Big2Str(bal),
	}, nil
}

func (r *RESTServer) GetUTXOs(c *gin.Context) (interface{}, error) {
	addrStr, exists := c.Params.Get("address")
	if !exists {
		return nil, errors.New("address is required")
	}
	addr := common.HexToAddress(addrStr)
	txs, err := r.storage.SpendableTxs(addr)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (r *RESTServer) GetBlock(c *gin.Context) (interface{}, error) {
	heightStr, exists := c.Params.Get("height")
	if !exists {
		return nil, errors.New("height is required")
	}
	height, ok := util.Str2Uint64(heightStr)
	if !ok {
		return nil, errors.New("invalid height")
	}
	block, meta, txs, err := r.storage.FullBlockAtHeight(height)
	if err != nil {
		return nil, err
	}

	return &gin.H{
		"block":                 block,
		"metadata":              meta,
		"confirmedTransactions": txs,
	}, nil
}

func (r *RESTServer) Send(c *gin.Context) (interface{}, error) {
	var tx chain.Transaction
	if err := c.ShouldBindJSON(&tx); err != nil {
		return nil, err
	}

	inclusion := r.mpool.Append(tx)
	if inclusion.Error != nil {
		return nil, inclusion.Error
	}

	tx.Body.BlockNumber = inclusion.BlockNumber
	tx.Body.TransactionIndex = inclusion.TransactionIndex
	return &gin.H{
		"transaction": tx,
		"inclusion":   inclusion,
	}, nil
}

func (r *RESTServer) Confirm(c *gin.Context) (interface{}, error) {
	var confirmation ConfirmationRequest
	if err := c.ShouldBindJSON(&confirmation); err != nil {
		return nil, err
	}
	sig0Bytes, err := hexutil.Decode(confirmation.ConfirmSig0)
	if err != nil {
		return nil, err
	}
	sig1Bytes, err := hexutil.Decode(confirmation.ConfirmSig1)
	if err != nil {
		return nil, err
	}

	var sig0 chain.Signature
	copy(sig0[:], sig0Bytes)
	var sig1 chain.Signature
	copy(sig1[:], sig1Bytes)

	tx, err := r.confirmer.Confirm(confirmation.BlockNumber, confirmation.TransactionIndex, [2]chain.Signature{
		sig0,
		sig1,
	})
	if err != nil {
		return nil, err
	}

	return tx, err
}

func (r *RESTServer) wrapHandler(f WrappableHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := f(c)
		if err != nil {
			log.WithError(restLogger, err).Error("received error in URL handler")
			c.Status(http.StatusBadRequest)
		} else {
			c.JSON(http.StatusOK, res)
		}
	}
}

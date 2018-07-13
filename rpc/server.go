package rpc

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	grpc "github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/node"
)

func Start(
	port int,
	storage db.PlasmaStorage,
	sink *node.TransactionSink,
) {
	log.Printf("Starting RPC server on port %d.\n", port)

	chch := make(chan chan node.TransactionRequest)

	txService := &TransactionService{
		TxChan: chch,
	}

	blockService := &BlockService{
		Storage: storage,
	}

	sink.AcceptTransactionRequests(chch)

	s := grpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterCodec(json.NewCodec(), "application/json;charset=utf-8")
	s.RegisterService(txService, "Transaction")
	s.RegisterService(blockService, "Block")
	r := mux.NewRouter()
	r.Handle("/rpc", s)
	http.ListenAndServe(fmt.Sprint(":", port), r)
}

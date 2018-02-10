package rpc

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/kyokan/plasma/node"
	"log"
	"net/http"
)

func Start(port int, txChan chan<- chan node.TransactionRequest) {
	log.Printf("Starting RPC server on port %d.", port)

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterCodec(json.NewCodec(), "application/json;charset=utf-8")
	s.RegisterService(&TransactionService{TxChan: txChan}, "Transaction")
	r := mux.NewRouter()
	r.Handle("/rpc", s)
	http.ListenAndServe(fmt.Sprint(":", port), r)
}

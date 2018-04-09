package validator

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/rpc"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/json"
)

func Run(port int) {
	fmt.Println("Validator Server Starting")
	log.Printf("Starting RPC server on port %d.", port)

	// TODO: we need to continually polling.

	go watchRootNode()
	go watchPlasmaContract()

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterCodec(json.NewCodec(), "application/json;charset=utf-8")
	s.RegisterService(&ValidatorService{}, "Validator")
	r := mux.NewRouter()
	r.Handle("/rpc", s)
	http.ListenAndServe(fmt.Sprint(":", port), r)
}

func watchRootNode() {
	for {
		fmt.Println("Watching root node...")
		time.Sleep(3 * time.Second)
	}
}

func watchPlasmaContract() {
	for {
		fmt.Println("Watching plasma contract...")
		time.Sleep(3 * time.Second)
	}
}

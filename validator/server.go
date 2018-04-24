package validator

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/rpc"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/json"
)

func Run(validatorPort int) {
	log.Printf("Starting validator server on port %d.", validatorPort)

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterCodec(json.NewCodec(), "application/json;charset=utf-8")
	s.RegisterService(&ValidatorService{}, "Validator")
	r := mux.NewRouter()
	r.Handle("/rpc", s)
	http.ListenAndServe(fmt.Sprint(":", validatorPort), r)
}

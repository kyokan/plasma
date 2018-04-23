package validator

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/rpc"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/json"
)

// TODO: clean up these args
func Run(rootPort int, validatorPort int, level *db.Database, plasma *eth.PlasmaClient) {
	log.Println("Validator Server Starting")
	log.Printf("Starting validator server on port %d.", validatorPort)

	// TODO: move it to the starter.
	go RootNodeListener(rootPort, level)
	go ExitStartedListener(rootPort, level, plasma)

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterCodec(json.NewCodec(), "application/json;charset=utf-8")
	s.RegisterService(&ValidatorService{}, "Validator")
	r := mux.NewRouter()
	r.Handle("/rpc", s)
	http.ListenAndServe(fmt.Sprint(":", validatorPort), r)
}

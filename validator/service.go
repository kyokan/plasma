package validator

import (
	"log"
	"net/http"
)

type StatusArgs struct {
	From string
}

type StatusResponse struct {
	Status string
}

type ValidatorService struct {
}

func (t *ValidatorService) Status(r *http.Request, args *StatusArgs, reply *StatusResponse) error {
	log.Printf("Received ValidatorService.Status request.")

	return nil
}

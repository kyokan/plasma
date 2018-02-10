package db

import (
	"log"
	"os/user"
	"path"
)

func DefaultLocation() string {
	usr, err := user.Current()

	if err != nil {
		log.Fatal(err)
	}

	return path.Join(usr.HomeDir, ".plasma")
}

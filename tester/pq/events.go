package pq

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/kyokan/plasma/contracts/gen/contracts"
)

func DebugUintFilter(pq *contracts.PriorityQueue) {
	opts := bind.FilterOpts{
		Start:   0x0, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := pq.FilterDebugUint(&opts)

	if err != nil {
		panic(err)
	}

	next := true
	length := 0

	for next {
		if itr.Event != nil {
			fmt.Println(itr.Event.Item)
			length++
		}
		next = itr.Next()
	}

	fmt.Printf("%d Debug Uint logs found \n", length)
}

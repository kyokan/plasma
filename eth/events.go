package eth

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/kyokan/plasma/eth/contracts"
)

func (c *clientState) filterOpts(start uint64) (*bind.FilterOpts, error) {
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	end := header.Number.Uint64()

	return &bind.FilterOpts{
		Start:   start,
		End:     &end, // TODO: end doesn't seem to work
		Context: context.Background(),
	}, nil
}

func (c *clientState) DepositFilter(start uint64, end uint64) ([]contracts.PlasmaDeposit, uint64, error) {
	opts, err := c.filterOpts(start)
	if err != nil {
		return nil, 0, err
	}
	opts.End = &end


	itr, err := c.contract.FilterDeposit(opts)
	if err != nil {
		log.Fatalf("Failed to filter deposit events: %v", err)
	}

	next := true
	var events []contracts.PlasmaDeposit
	for next {
		if itr.Event != nil {
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, end, nil
}

func (c *clientState) ChallengedExitFilter(start uint64) ([]contracts.PlasmaChallengedExit, uint64, error) {
	opts, err := c.filterOpts(start)
	if err != nil {
		return nil, 0, err
	}
	end := *opts.End


	itr, err := c.contract.FilterChallengedExit(opts)
	if err != nil {
		log.Fatalf("Failed to filter challenged exit events: %v", err)
	}

	next := true
	var events []contracts.PlasmaChallengedExit
	for next {
		if itr.Event != nil {
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, end, nil
}


func (c *clientState) FinalizedExitFilter(start uint64) ([]contracts.PlasmaFinalizedExit, uint64, error) {
	opts, err := c.filterOpts(start)
	if err != nil {
		return nil, 0, err
	}
	end := *opts.End


	itr, err := c.contract.FilterFinalizedExit(opts)
	if err != nil {
		log.Fatalf("Failed to filter finalized exit events: %v", err)
	}

	next := true
	var events []contracts.PlasmaFinalizedExit
	for next {
		if itr.Event != nil {
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, end, nil
}

func (c *clientState) StartedTransactionExitFilter(start uint64) ([]contracts.PlasmaStartedTransactionExit, uint64, error) {
	opts, err := c.filterOpts(start)
	if err != nil {
		return nil, 0, err
	}
	end := *opts.End


	itr, err := c.contract.FilterStartedTransactionExit(opts)
	if err != nil {
		log.Fatalf("Failed to filter started transaction exit events: %v", err)
	}

	next := true
	var events []contracts.PlasmaStartedTransactionExit
	for next {
		if itr.Event != nil {
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, end, nil
}

func (c *clientState) StartedDepositExitFilter(start uint64) ([]contracts.PlasmaStartedDepositExit, uint64, error) {
	opts, err := c.filterOpts(start)
	if err != nil {
		return nil, 0, err
	}
	end := *opts.End


	itr, err := c.contract.FilterStartedDepositExit(opts)
	if err != nil {
		log.Fatalf("Failed to filter started deposit exit events: %v", err)
	}

	next := true
	var events []contracts.PlasmaStartedDepositExit
	for next {
		if itr.Event != nil {
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, end, nil
}


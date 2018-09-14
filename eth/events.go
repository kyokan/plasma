package eth

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/kyokan/plasma/contracts/gen/contracts"
)

func (c *clientState) DepositFilter(start uint64) ([]contracts.PlasmaDeposit, uint64, error) {
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, 0, err
	}

	end := header.Number.Uint64()

	opts := bind.FilterOpts{
		Start:   start,
		End:     &end, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := c.contract.FilterDeposit(&opts)
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

func (c *clientState) ExitStartedFilter(start uint64) ([]contracts.PlasmaExitStarted, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := c.contract.FilterExitStarted(&opts)

	if err != nil {
		log.Fatalf("Failed to filter exit started events: %v", err)
	}

	next := true

	var events []contracts.PlasmaExitStarted

	var lastBlockNumber uint64

	for next {
		if itr.Event != nil {
			lastBlockNumber = itr.Event.Raw.BlockNumber
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, lastBlockNumber
}

func (c *clientState) DebugAddressFilter(start uint64) ([]contracts.PlasmaDebugAddress, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := c.contract.FilterDebugAddress(&opts)

	if err != nil {
		log.Fatalf("Failed to filter debug address events: %v", err)
	}

	next := true

	var events []contracts.PlasmaDebugAddress

	var lastBlockNumber uint64

	for next {
		if itr.Event != nil {
			lastBlockNumber = itr.Event.Raw.BlockNumber
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, lastBlockNumber
}

func (c *clientState) DebugUintFilter(start uint64) ([]contracts.PlasmaDebugUint, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := c.contract.FilterDebugUint(&opts)

	if err != nil {
		log.Fatalf("Failed to filter debug uint events: %v", err)
	}

	next := true

	var events []contracts.PlasmaDebugUint

	var lastBlockNumber uint64

	for next {
		if itr.Event != nil {
			lastBlockNumber = itr.Event.Raw.BlockNumber
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, lastBlockNumber
}

func (c *clientState) DebugBoolFilter(start uint64) ([]contracts.PlasmaDebugBool, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := c.contract.FilterDebugBool(&opts)

	if err != nil {
		log.Fatalf("Failed to filter debug bool events: %v", err)
	}

	next := true

	var events []contracts.PlasmaDebugBool

	var lastBlockNumber uint64

	for next {
		if itr.Event != nil {
			lastBlockNumber = itr.Event.Raw.BlockNumber
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, lastBlockNumber
}

func (c *clientState) ChallengeSuccessFilter(
	start uint64,
) ([]contracts.PlasmaChallengeSuccess, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := c.contract.FilterChallengeSuccess(&opts)

	if err != nil {
		log.Fatalf("Failed to filter challenge success events: %v", err)
	}

	next := true

	var events []contracts.PlasmaChallengeSuccess

	var lastBlockNumber uint64

	for next {
		if itr.Event != nil {
			lastBlockNumber = itr.Event.Raw.BlockNumber
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, lastBlockNumber
}

func (c *clientState) ChallengeFailureFilter(
	start uint64,
) ([]contracts.PlasmaChallengeFailure, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := c.contract.FilterChallengeFailure(&opts)

	if err != nil {
		log.Fatalf("Failed to filter challenge failure events: %v", err)
	}

	next := true

	var events []contracts.PlasmaChallengeFailure

	var lastBlockNumber uint64

	for next {
		if itr.Event != nil {
			lastBlockNumber = itr.Event.Raw.BlockNumber
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, lastBlockNumber
}

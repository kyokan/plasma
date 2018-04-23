package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/kyokan/plasma/contracts/gen/contracts"
)

func (p *PlasmaClient) DepositFilter(
	start uint64,
) ([]contracts.PlasmaDeposit, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterDeposit(&opts)

	if err != nil {
		panic(err)
	}

	next := true

	var events []contracts.PlasmaDeposit

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

func (p *PlasmaClient) ExitStartedFilter(
	start uint64,
) ([]contracts.PlasmaExitStarted, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterExitStarted(&opts)

	if err != nil {
		panic(err)
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

func (p *PlasmaClient) DebugAddressFilter(
	start uint64,
) ([]contracts.PlasmaDebugAddress, uint64) {
	opts := bind.FilterOpts{
		Start:   start, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterDebugAddress(&opts)

	if err != nil {
		panic(err)
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

func (p *PlasmaClient) DebugUintFilter(
	start uint64,
) ([]contracts.PlasmaDebugUint, uint64) {
	opts := bind.FilterOpts{
		Start:   start, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterDebugUint(&opts)

	if err != nil {
		panic(err)
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

func (p *PlasmaClient) DebugBoolFilter(
	start uint64,
) ([]contracts.PlasmaDebugBool, uint64) {
	opts := bind.FilterOpts{
		Start:   start, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterDebugBool(&opts)

	if err != nil {
		panic(err)
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

func (p *PlasmaClient) ChallengeSuccessFilter(
	start uint64,
) ([]contracts.PlasmaChallengeSuccess, uint64) {
	opts := bind.FilterOpts{
		Start:   0x0, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterChallengeSuccess(&opts)

	if err != nil {
		panic(err)
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

func (p *PlasmaClient) ChallengeFailureFilter(
	start uint64,
) ([]contracts.PlasmaChallengeFailure, uint64) {
	opts := bind.FilterOpts{
		Start:   0x0, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterChallengeFailure(&opts)

	if err != nil {
		panic(err)
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

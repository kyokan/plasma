package service

import (
	"github.com/kyokan/plasma/pkg/db"
	"github.com/kyokan/plasma/pkg/rpc/pb"
	"context"
	"time"
	"io"
	"github.com/kyokan/plasma/pkg/chain"
	"github.com/kyokan/plasma/pkg/validation"
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/pkg/log"
	"github.com/sirupsen/logrus"
)

type Syncer struct {
	storage    db.Storage
	rootClient pb.RootClient
	ethClient  eth.Client
	quit       chan bool
}

const syncerPollInterval = time.Second

var syncerLogger = log.ForSubsystem("Syncer")

func NewSyncer(storage db.Storage, rootClient pb.RootClient, ethClient eth.Client) *Syncer {
	return &Syncer{
		storage:    storage,
		rootClient: rootClient,
		ethClient:  ethClient,
		quit:       make(chan bool),
	}
}

func (s *Syncer) Start() error {
	go func() {
		syncerLogger.Info("started sync process")

		for {
			select {
			case <-s.quit:
				return
			default:
				start := time.Now()
				if err := s.doSync(); err != nil {
					log.WithError(syncerLogger, err).Error("encountered error during sync")
				}
				elapsed := time.Since(start)
				if elapsed < syncerPollInterval {
					time.Sleep(syncerPollInterval - elapsed)
				}
			}
		}
	}()

	return nil
}

func (s *Syncer) Stop() error {
	s.quit <- true
	return nil
}

func (s *Syncer) doSync() error {
	head, err := s.storage.LatestBlock()
	if err != nil {
		return err
	}

	start := uint64(1)
	if head != nil {
		start = head.Header.Number + 1
	}
	syncerLogger.WithFields(logrus.Fields{
		"start": start,
	}).Info("started sync")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := s.rootClient.Sync(ctx, &pb.SyncRequest{
		Start: start,
	})
	if err != nil {
		return err
	}

	end := start
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		confirmedTxs := make([]chain.ConfirmedTransaction, len(res.ConfirmedTransactions), len(res.ConfirmedTransactions))
		for i, protoTx := range res.ConfirmedTransactions {
			tx, err := chain.ConfirmedTransactionFromProto(protoTx)
			if err != nil {
				return err
			}
			confirmedTxs[i] = *tx
		}
		block := chain.BlockFromProto(res.Block)
		meta := chain.BlockMetadataFromProto(res.Metadata)
		if err := validation.ValidateBlock(s.storage, s.ethClient, block, confirmedTxs); err != nil {
			return err
		}
		if err := s.storage.InsertBlock(block, meta, confirmedTxs); err != nil {
			return nil
		}
		end = block.Header.Number
	}

	syncerLogger.WithFields(logrus.Fields{
		"imported": end - start,
		"start":    start,
		"end":      end,
	}).Info("finished sync")

	return nil
}

package synchronizer

import (
	"context"
	"sync"
	"time"

	"github.com/m1khal3v/gophkeeper/internal/client/grpc"
	"github.com/m1khal3v/gophkeeper/internal/client/manager"
	"github.com/m1khal3v/gophkeeper/internal/client/model"
	"github.com/m1khal3v/gophkeeper/internal/common/logger"
	"go.uber.org/zap"
)

type Synchronizer struct {
	client      *grpc.Client
	userDataMgr *manager.UserDataManager
	metaManager *manager.MetaManager
	interval    time.Duration
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

func New(
	client *grpc.Client,
	userDataMgr *manager.UserDataManager,
	metaManager *manager.MetaManager,
	interval time.Duration,
) *Synchronizer {
	return &Synchronizer{
		client:      client,
		userDataMgr: userDataMgr,
		metaManager: metaManager,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

func (s *Synchronizer) Start(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.syncOnce(ctx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.syncOnce(ctx)
			case <-s.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Synchronizer) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

func (s *Synchronizer) syncOnce(ctx context.Context) {
	lastSync, err := s.metaManager.GetLastSync(ctx)
	if err != nil {
		logger.Logger.Fatal("sync: get lastSync error:", zap.Error(err))
	}

	if s.pushLocalUpdates(ctx, lastSync) && s.fetchRemoteUpdates(ctx, lastSync) {
		if err := s.metaManager.SetLastSync(ctx, time.Now().UTC()); err != nil {
			logger.Logger.Fatal("sync: set lastSync error:", zap.Error(err))
		}
	}
}

func (s *Synchronizer) pushLocalUpdates(ctx context.Context, lastSync time.Time) bool {
	localUpdates, err := s.userDataMgr.GetUpdates(ctx, lastSync)
	if err != nil {
		logger.Logger.Fatal("sync: can't get local updates:", zap.Error(err))
	}

	for _, data := range localUpdates {
		_, err := s.client.Upsert(ctx, data)
		if err != nil {
			logger.Logger.Warn("sync: can't push local update to server", zap.Error(err))

			return false
		}
	}

	return true
}

func (s *Synchronizer) fetchRemoteUpdates(ctx context.Context, lastSync time.Time) bool {
	resp, err := s.client.GetUpdates(ctx, lastSync)
	if err != nil {
		logger.Logger.Warn("sync: can`t get updates:", zap.Error(err))

		return false
	}

	for _, item := range resp.Items {
		data := &model.UserData{
			DataKey:   item.DataKey,
			DataValue: item.DataValue,
			UpdatedAt: item.UpdatedAt.AsTime(),
			DeletedAt: item.DeletedAt.AsTime(),
		}

		if err := s.userDataMgr.Upsert(ctx, data); err != nil {
			logger.Logger.Fatal("sync: can't update local data:", zap.Error(err))
		}
	}

	return true
}

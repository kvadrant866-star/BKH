package service

import (
	"context"
	"time"

	"banner/internal/agregator"
	"banner/internal/entity"
)

type Delta = agregator.Delta

type BannerRepositorier interface {
	UpsertBatch(ctx context.Context, deltas []Delta) error
	SelectStats(ctx context.Context, bannerID int, from, to time.Time) ([]entity.MinuteStat, error)
}

type BannerService struct {
	repo   BannerRepositorier
	agg    *agregator.Agregator
	ticker *time.Ticker
	stop   chan struct{}
}

func NewBannerService(repo BannerRepositorier, numShards int, flushInterval time.Duration) *BannerService {
	s := &BannerService{repo: repo, agg: agregator.NewAgregator(numShards), stop: make(chan struct{})}
	s.ticker = time.NewTicker(flushInterval)
	go s.flushLoop()
	return s
}

func (s *BannerService) IncrementClick(ctx context.Context, bannerID int) error {
	now := time.Now().UTC()
	t := now.Truncate(time.Minute)
	s.agg.Increment(bannerID, t)
	//log.Printf("click banner_id=%d ts=%s", bannerID, now.Format("2006-01-02T15:04:05"))
	return nil
}

func (s *BannerService) GetStats(ctx context.Context, bannerID int, from, to time.Time) ([]entity.MinuteStat, error) {
	return s.repo.SelectStats(ctx, bannerID, from, to)
}

func (s *BannerService) flushLoop() {
	for {
		select {
		case <-s.ticker.C:
			deltas := s.agg.FlushDeltas()
			if len(deltas) == 0 {
				continue
			}
			_ = s.repo.UpsertBatch(context.Background(), deltas)
		case <-s.stop:
			deltas := s.agg.FlushDeltas()
			if len(deltas) > 0 {
				_ = s.repo.UpsertBatch(context.Background(), deltas)
			}
			s.ticker.Stop()
			return
		}
	}
}

func (s *BannerService) Close() { close(s.stop) }

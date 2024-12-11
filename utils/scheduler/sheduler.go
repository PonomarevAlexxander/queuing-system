package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"go.uber.org/zap"

	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
)

type ScheduledTask func(context.Context) error

type BackoffGetter interface {
	NextInterval() time.Duration
}

type Scheduler struct {
	log     *logger.Logger
	clk     clock.Clock
	stopped chan struct{}
}

func NewScheduler(log *logger.Logger, clk clock.Clock) *Scheduler {
	return &Scheduler{
		log:     log,
		clk:     clk,
		stopped: make(chan struct{}),
	}
}

// Run blocks until scheduler stopps
func (s *Scheduler) Run(ctx context.Context, backoff BackoffGetter, task ScheduledTask) {
	ticker := s.clk.Ticker(backoff.NextInterval())
	defer ticker.Stop()
	var wg sync.WaitGroup

outer:
	for {
		select {
		case <-ticker.C:
			ticker.Reset(backoff.NextInterval())
			wg.Add(1)
			go func() {
				defer s.log.LogPanic()
				defer wg.Done()

				if err := task(ctx); err != nil {
					s.log.Error("Scheduled task failed with error", zap.Error(err))
				}
			}()
		case <-ctx.Done():
			s.log.Warn(
				"Scheduler stopped ucexpectedly, wait tasks complete and exit",
				zap.Error(ctx.Err()),
			)
			break outer
		case <-s.stopped:
			s.log.Info("Stopping scheduler gracefully...")
			break outer
		}
	}

	wg.Wait()
}

func (s *Scheduler) Stop() {
	close(s.stopped)
}

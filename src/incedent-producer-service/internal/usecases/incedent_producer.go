package usecases

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/benbjohnson/clock"
	"go.uber.org/zap"

	"github.com/PonomarevAlexxander/queuing-system/incedent-producer-service/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
	"github.com/PonomarevAlexxander/queuing-system/utils/scheduler"
)

type dispatcherClient interface {
	SendIncedent(ctx context.Context, incedent domain.Incedent) error
}

type configStorage interface {
	GetConfig()
}

type scheduledRunner interface {
	Run(ctx context.Context, backoff scheduler.BackoffGetter, task scheduler.ScheduledTask)
	Stop()
}

type IncedentProducer struct {
	log *logger.Logger
	clk clock.Clock

	client dispatcherClient
	runner scheduledRunner

	interval time.Duration
	counter  atomic.Uint64
	priority domain.Priority
}

func NewIncedentProducer(
	log *logger.Logger,
	clk clock.Clock,
	client dispatcherClient,
	runner scheduledRunner,
	interval time.Duration,
	priority domain.Priority,
) *IncedentProducer {
	return &IncedentProducer{
		log:      log,
		clk:      clk,
		client:   client,
		runner:   runner,
		priority: priority,
		interval: interval,
	}
}

func (ip *IncedentProducer) Run(ctx context.Context) error {
	ip.runner.Run(ctx, scheduler.NewLinearBackoff(ip.interval), ip.generateIncedent)

	return nil
}

func (ip *IncedentProducer) generateIncedent(ctx context.Context) error {
	incedent := domain.Incedent{
		Id:           ip.counter.Add(1),
		CreationTime: ip.clk.Now(),
		Priority:     ip.priority,
	}
	if err := ip.client.SendIncedent(ctx, incedent); err != nil {
		return fmt.Errorf("failed to send incedent: %w", err)
	}
	ip.log.Debug(
		"New incedent generated",
		zap.Stringer("Incedent", incedent),
	)
	return nil
}

func (ip *IncedentProducer) Stop() {
	ip.runner.Stop()
}

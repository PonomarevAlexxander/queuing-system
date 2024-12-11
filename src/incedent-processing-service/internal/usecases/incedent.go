package usecases

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"go.uber.org/zap"

	"github.com/PonomarevAlexxander/queuing-system/incedent-processing-service/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
)

type backoffGetter interface {
	NextInterval() time.Duration
}

type incedentProcessingUseCase struct {
	log     *logger.Logger
	clk     clock.Clock
	backoff backoffGetter
}

func NewIncedentProcessingUseCase(
	log *logger.Logger,
	clk clock.Clock,
	backoff backoffGetter,
) *incedentProcessingUseCase {
	return &incedentProcessingUseCase{
		log:     log,
		clk:     clk,
		backoff: backoff,
	}
}

func (ip *incedentProcessingUseCase) ProcessIncedent(ctx context.Context, incedent domain.Incedent) error {
	interval := ip.backoff.NextInterval()
	ip.log.Info(
		"New incedent received, start processing",
		zap.Stringer("incedent", incedent),
		zap.Duration("interval", interval),
	)
	timer := ip.clk.Timer(interval)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		ip.log.Info("Incedent processed", zap.Stringer("incedent", incedent))
		return nil
	}
}

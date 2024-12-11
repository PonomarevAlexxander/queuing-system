package usecases

import (
	"context"
	"fmt"

	"github.com/PonomarevAlexxander/queuing-system/incedent-processing-service/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
	"github.com/benbjohnson/clock"
	"go.uber.org/zap"
)

const (
	triesNumber = 5
)

type registerClient interface {
	Register(ctx context.Context, info domain.RegistrationInfo) error
}

type registerUseCase struct {
	log     *logger.Logger
	clk     clock.Clock
	regInfo domain.RegistrationInfo
	client  registerClient
}

func NewRegisterUseCase(
	log *logger.Logger, clk clock.Clock,
	regInfo domain.RegistrationInfo, regClient registerClient,
) *registerUseCase {
	return &registerUseCase{
		log:     log,
		clk:     clk,
		regInfo: regInfo,
		client:  regClient,
	}
}

func (r *registerUseCase) Run(ctx context.Context) error {
	return r.tryRegister(ctx)
}

func (r *registerUseCase) Stop() {
}

func (r *registerUseCase) tryRegister(ctx context.Context) error {
	var err error
	for counter := 0; counter < triesNumber; counter++ {
		if err = r.client.Register(ctx, r.regInfo); err == nil {
			r.log.Info("Successfully registered in dispatcher")
			return nil
		}
		r.log.Debug("Failed to register, trying one more time", zap.Error(err))
	}

	return fmt.Errorf("failed to register: %w", err)
}

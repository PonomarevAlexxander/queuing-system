package runner

import (
	"context"
	"os"
	"os/signal"
	"reflect"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
)

type Service interface {
	Run(ctx context.Context) error
	Stop()
}

type ServiceRunner struct {
	log *logger.Logger
}

func NewServiceRunner(log *logger.Logger, signals ...os.Signal) (context.Context, context.CancelFunc, *ServiceRunner) {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, signals...)
	return ctx, stop, &ServiceRunner{
		log: log,
	}
}

func (sr *ServiceRunner) Run(ctx context.Context, services ...Service) {
	group, ctx := errgroup.WithContext(ctx)
	for _, service := range services {
		sr.log.Info("Starting new service", zap.String("<service>", getType(service)))
		group.Go(func() error {
			return service.Run(ctx)
		})
	}

	if err := group.Wait(); err != nil {
		sr.log.Error("Exiting from runner with error", zap.Error(err))
	}

	var eg errgroup.Group
	for _, service := range services {
		sr.log.Info("Stopping service", zap.String("<service>", getType(service)))
		eg.Go(func() error {
			service.Stop()
			sr.log.Info("Stopped service", zap.String("<service>", getType(service)))
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		sr.log.Error("Exiting from runner with error", zap.Error(err))
	}

	sr.log.Info("All services stoped, exiting runner")
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

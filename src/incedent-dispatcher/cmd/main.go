package main

import (
	"fmt"
	"net"
	"syscall"

	"github.com/alexflint/go-arg"
	"github.com/benbjohnson/clock"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/config"
	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/controllers"
	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/repositories"
	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/usecases"
	"github.com/PonomarevAlexxander/queuing-system/services/incedent_dispatcher"
	common_config "github.com/PonomarevAlexxander/queuing-system/utils/config"
	grpc_controller "github.com/PonomarevAlexxander/queuing-system/utils/grpc_controller"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
	"github.com/PonomarevAlexxander/queuing-system/utils/runner"
)

var args struct {
	Config string `arg:"required"`
}

func main() {
	arg.MustParse(&args)

	cfg, err := common_config.ReadConfigFromYAML[config.DispatcherConfig](args.Config)
	if err != nil {
		panic(err)
	}

	err = common_config.ValidateConfig(cfg)
	if err != nil {
		panic(err)
	}

	zapLog, err := logger.InitZapLogger(cfg.LoggerConfig)
	if err != nil {
		panic(err)
	}
	log := logger.InitZapWrapper(zapLog)

	ctx, cancel, srvcRunner := runner.NewServiceRunner(log, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	clk := clock.New()
	bfStorage := repositories.NewBufferStorage(log, cfg.InnerConfig.BufferCapacity)
	procStorage := repositories.NewProcessorStorage()
	mStorage := repositories.NewMetricsStorage(log, clk)
	registrationUC := usecases.NewRegistrationUseCase(log, procStorage, mStorage)
	dispatcherUC := usecases.NewIncedentDispatcher(log, clk, bfStorage, procStorage, mStorage)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", cfg.InnerConfig.Port))
	if err != nil {
		log.Fatal("Failed to create tcp server", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	dispatcherController := controllers.NewGrpcController(log, registrationUC, dispatcherUC)
	incedent_dispatcher.RegisterIncedentDispatcherServer(grpcServer, dispatcherController)
	controller := grpc_controller.NewGrpcController(grpcServer, lis)

	srvcRunner.Run(ctx, registrationUC, dispatcherUC, controller)
	mStorage.PrintStatistics()
}

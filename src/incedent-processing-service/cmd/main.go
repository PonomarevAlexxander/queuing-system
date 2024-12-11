package main

import (
	"fmt"
	"net"
	"syscall"

	"github.com/alexflint/go-arg"
	"github.com/benbjohnson/clock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/PonomarevAlexxander/queuing-system/incedent-processing-service/internal/clients"
	"github.com/PonomarevAlexxander/queuing-system/incedent-processing-service/internal/config"
	"github.com/PonomarevAlexxander/queuing-system/incedent-processing-service/internal/controllers"
	"github.com/PonomarevAlexxander/queuing-system/incedent-processing-service/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/incedent-processing-service/internal/usecases"
	"github.com/PonomarevAlexxander/queuing-system/services/incedent_dispatcher"
	"github.com/PonomarevAlexxander/queuing-system/services/incedent_processor"
	common_config "github.com/PonomarevAlexxander/queuing-system/utils/config"
	grpc_controller "github.com/PonomarevAlexxander/queuing-system/utils/grpc_controller"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
	"github.com/PonomarevAlexxander/queuing-system/utils/runner"
	"github.com/PonomarevAlexxander/queuing-system/utils/scheduler"
)

var args struct {
	Id     uint64 `arg:"required"`
	Host   string `arg:"required"`
	Config string `arg:"required"`
}

func main() {
	arg.MustParse(&args)

	cfg, err := common_config.ReadConfigFromYAML[config.IncedentProcessorConfig](args.Config)
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

	conn, err := grpc.NewClient(cfg.DispatcherConfig.Host,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to create grpc client", zap.Error(err))
	}
	defer conn.Close()
	client := incedent_dispatcher.NewIncedentDispatcherClient(conn)
	regClient := clients.NewRegisterClient(client)

	clk := clock.New()
	processingUC := usecases.NewIncedentProcessingUseCase(log, clk,
		scheduler.NewExponentialBackoff(cfg.InnerConfig.GetInterval()))
	registerUC := usecases.NewRegisterUseCase(
		log,
		clk,
		domain.RegistrationInfo{
			Id:   args.Id,
			Host: args.Host,
		},
		regClient,
	)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", config.GetPort(args.Host)))
	if err != nil {
		log.Fatal("Failed to create tcp server", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	processingController := controllers.NewGrpcController(log, processingUC)
	incedent_processor.RegisterIncedentProcessorServer(grpcServer, processingController)
	controller := grpc_controller.NewGrpcController(grpcServer, lis)

	srvcRunner.Run(ctx, controller, registerUC)
}

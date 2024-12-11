package main

import (
	"syscall"

	arg "github.com/alexflint/go-arg"
	"github.com/benbjohnson/clock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/PonomarevAlexxander/queuing-system/incedent-producer-service/internal/clients"
	"github.com/PonomarevAlexxander/queuing-system/incedent-producer-service/internal/config"
	"github.com/PonomarevAlexxander/queuing-system/incedent-producer-service/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/incedent-producer-service/internal/usecases"
	"github.com/PonomarevAlexxander/queuing-system/services/incedent_dispatcher"
	common_config "github.com/PonomarevAlexxander/queuing-system/utils/config"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
	"github.com/PonomarevAlexxander/queuing-system/utils/runner"
	"github.com/PonomarevAlexxander/queuing-system/utils/scheduler"
)

var args struct {
	Config   string `arg:"required"`
	Priority uint64 `arg:"required"`
}

func main() {
	arg.MustParse(&args)

	cfg, err := common_config.ReadConfigFromYAML[config.IncedentProducerConfig](args.Config)
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

	ctx, cancel, srvcRunner := runner.NewServiceRunner(log,
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conn, err := grpc.NewClient(cfg.DispatcherConfig.Host,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to create grpc client", zap.Error(err))
	}
	defer conn.Close()
	grpcClient := incedent_dispatcher.NewIncedentDispatcherClient(conn)
	dispatcherClient := clients.NewDispatcherClient(grpcClient)
	clk := clock.New()
	schdlrRunner := scheduler.NewScheduler(log, clk)
	producer := usecases.NewIncedentProducer(
		log,
		clk,
		dispatcherClient,
		schdlrRunner,
		cfg.InnerConfig.GetInterval(),
		domain.Priority(args.Priority),
	)

	srvcRunner.Run(ctx, producer)
}

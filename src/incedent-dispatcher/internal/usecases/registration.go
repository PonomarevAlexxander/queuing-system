package usecases

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/clients"
	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/services/incedent_processor"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
)

type closeConnection func() error

type RegistrationUseCase struct {
	log               *logger.Logger
	processorsStorage processorsStorage
	metricsStorage    metricsStorage

	mu          sync.Mutex
	connections []closeConnection
}

func NewRegistrationUseCase(
	log *logger.Logger,
	processorsStorage processorsStorage,
	metricsStorage metricsStorage,
) *RegistrationUseCase {
	return &RegistrationUseCase{
		log:               log,
		processorsStorage: processorsStorage,
		metricsStorage:    metricsStorage,
		connections:       make([]closeConnection, 0),
	}
}

func (ru *RegistrationUseCase) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (ru *RegistrationUseCase) Stop() {
	ru.closeConnections()
}

func (ru *RegistrationUseCase) Register(ctx context.Context, processor domain.IncedentProcessor) error {
	client, err := ru.createClient(processor.Host)
	if err != nil {
		return err
	}

	ru.metricsStorage.RegisteredProcessor(processor)
	clientInfo := domain.ProcessorClientInfo{
		Processor: processor,
		Client:    client,
	}
	ru.processorsStorage.Add(clientInfo)
	ru.log.Info("New processor registered", zap.Stringer("processor", processor))

	return nil
}

func (ru *RegistrationUseCase) createClient(host string) (*clients.ProcessorClient, error) {
	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}
	ru.saveConnection(conn.Close)
	client := incedent_processor.NewIncedentProcessorClient(conn)

	return clients.NewProcessorClient(client), nil
}

func (ru *RegistrationUseCase) saveConnection(conn closeConnection) {
	ru.mu.Lock()
	defer ru.mu.Unlock()

	ru.connections = append(ru.connections, conn)
}

func (ru *RegistrationUseCase) closeConnections() {
	ru.mu.Lock()
	defer ru.mu.Unlock()

	for _, conn := range ru.connections {
		if err := conn(); err != nil {
			ru.log.Error("Failed to close connection", zap.Error(err))
		}
	}
}

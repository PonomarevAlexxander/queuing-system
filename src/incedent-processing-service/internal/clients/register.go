package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/PonomarevAlexxander/queuing-system/incedent-processing-service/internal/domain"
	msgs_dispatcher "github.com/PonomarevAlexxander/queuing-system/messages/registration"
	srvc_dispatcher "github.com/PonomarevAlexxander/queuing-system/services/incedent_dispatcher"
)

const (
	requestTimeout time.Duration = 5 * time.Second
)

type RegisterClient struct {
	grpcClient srvc_dispatcher.IncedentDispatcherClient
}

func NewRegisterClient(
	grpcClient srvc_dispatcher.IncedentDispatcherClient,
) *RegisterClient {
	return &RegisterClient{
		grpcClient: grpcClient,
	}
}

func (dc *RegisterClient) Register(ctx context.Context, info domain.RegistrationInfo) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req := &msgs_dispatcher.ProcessorRegisterReq{
		Id:   info.Id,
		Host: info.Host,
	}

	resp, err := dc.grpcClient.RegisterProcessor(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send registration req with grpc: %w", err)
	}

	if !resp.Result.GetSuccess() {
		return fmt.Errorf("registration wasn't handled: %w", domain.ErrBadResult)
	}

	return nil
}

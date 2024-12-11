package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/PonomarevAlexxander/queuing-system/incedent-producer-service/internal/domain"
	msgs_dispatcher "github.com/PonomarevAlexxander/queuing-system/messages/incedent"
	srvc_dispatcher "github.com/PonomarevAlexxander/queuing-system/services/incedent_dispatcher"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	requestTimeout time.Duration = 5 * time.Second
)

type DispatcherClient struct {
	grpcClient srvc_dispatcher.IncedentDispatcherClient
}

func NewDispatcherClient(
	grpcClient srvc_dispatcher.IncedentDispatcherClient,
) *DispatcherClient {
	return &DispatcherClient{
		grpcClient: grpcClient,
	}
}

func (dc *DispatcherClient) SendIncedent(ctx context.Context, incedent domain.Incedent) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req := &msgs_dispatcher.NewIncedentReq{
		Id:       incedent.Id,
		Time:     timestamppb.New(incedent.CreationTime),
		Priority: uint64(incedent.Priority),
	}

	resp, err := dc.grpcClient.NewIncedent(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send incedent with grpc: %w", err)
	}

	if !resp.Result.GetSuccess() {
		return fmt.Errorf("incedent wasn't handled, '%s': %w", resp.Result.Msg, domain.ErrBadResult)
	}

	return nil
}

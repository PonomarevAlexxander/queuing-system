package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/domain"
	msgs_processor "github.com/PonomarevAlexxander/queuing-system/messages/incedent"
	srvc_processor "github.com/PonomarevAlexxander/queuing-system/services/incedent_processor"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	requestTimeout time.Duration = 5 * time.Second
)

type ProcessorClient struct {
	grpcClient srvc_processor.IncedentProcessorClient
}

func NewProcessorClient(
	grpcClient srvc_processor.IncedentProcessorClient,
) *ProcessorClient {
	return &ProcessorClient{
		grpcClient: grpcClient,
	}
}

func (dc *ProcessorClient) SendIncedent(ctx context.Context, incedent domain.Incedent) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req := &msgs_processor.NewIncedentReq{
		Id:       incedent.Id,
		Time:     timestamppb.New(incedent.CreationTime),
		Priority: uint64(incedent.Priority),
	}

	resp, err := dc.grpcClient.NewIncedent(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send incedent with grpc: %w", err)
	}

	if !resp.Result.GetSuccess() {
		return fmt.Errorf("incedent wasn't handled: %w", domain.ErrBadResult)
	}

	return nil
}

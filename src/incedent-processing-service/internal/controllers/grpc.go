package controllers

import (
	"context"

	"github.com/PonomarevAlexxander/queuing-system/incedent-processing-service/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/messages/common"
	"github.com/PonomarevAlexxander/queuing-system/messages/incedent"
	"github.com/PonomarevAlexxander/queuing-system/services/incedent_processor"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
)

type processorUC interface {
	ProcessIncedent(ctx context.Context, incedent domain.Incedent) error
}

type GrpcController struct {
	incedent_processor.UnimplementedIncedentProcessorServer
	log       *logger.Logger
	processor processorUC
}

func NewGrpcController(
	log *logger.Logger,
	processor processorUC,
) *GrpcController {
	return &GrpcController{
		log:       log,
		processor: processor,
	}
}

func (gc *GrpcController) NewIncedent(ctx context.Context, req *incedent.NewIncedentReq) (*incedent.NewIncedentResp, error) {
	resp := &incedent.NewIncedentResp{
		Result: &common.Result{
			Success: true,
		},
	}

	if err := gc.processor.ProcessIncedent(
		ctx,
		domain.Incedent{
			Id:           req.GetId(),
			Priority:     domain.Priority(req.GetPriority()),
			CreationTime: req.GetTime().AsTime(),
		},
	); err != nil {
		resp.Result.Success = false
		resp.Result.Msg = err.Error()
		return resp, nil
	}

	return resp, nil
}

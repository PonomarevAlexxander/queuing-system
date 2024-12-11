package controllers

import (
	"context"

	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/messages/common"
	"github.com/PonomarevAlexxander/queuing-system/messages/incedent"
	"github.com/PonomarevAlexxander/queuing-system/messages/registration"
	"github.com/PonomarevAlexxander/queuing-system/services/incedent_dispatcher"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
)

type registerUC interface {
	Register(ctx context.Context, processor domain.IncedentProcessor) error
}

type dispatcher interface {
	NewIncedent(ctx context.Context, incedent domain.Incedent) error
}

type GrpcController struct {
	incedent_dispatcher.UnimplementedIncedentDispatcherServer
	log        *logger.Logger
	registerUC registerUC
	dispatcher dispatcher
}

func NewGrpcController(
	log *logger.Logger,
	registerUC registerUC,
	dispatcherUC dispatcher,
) *GrpcController {
	return &GrpcController{
		log:        log,
		registerUC: registerUC,
		dispatcher: dispatcherUC,
	}
}

func (gc *GrpcController) NewIncedent(ctx context.Context, req *incedent.NewIncedentReq) (*incedent.NewIncedentResp, error) {
	resp := &incedent.NewIncedentResp{
		Result: &common.Result{
			Success: true,
		},
	}

	if err := gc.dispatcher.NewIncedent(
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

func (gc *GrpcController) RegisterProcessor(ctx context.Context, req *registration.ProcessorRegisterReq) (*registration.ProcessorRegisterResp, error) {
	resp := &registration.ProcessorRegisterResp{
		Result: &common.Result{
			Success: true,
		},
	}

	if err := gc.registerUC.Register(
		ctx,
		domain.IncedentProcessor{Id: req.GetId(), Host: req.GetHost()},
	); err != nil {
		resp.Result.Success = false
		resp.Result.Msg = err.Error()
		return resp, nil
	}

	return resp, nil
}

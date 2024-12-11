package usecases

import (
	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/domain"
)

type bufferStorage interface {
	CheckAndPut(incedent domain.Incedent) error
	DeleteIncedent(incedent domain.Incedent) error
	EvictAndPut(incedent domain.Incedent) domain.Incedent
	GetPacket() []domain.Incedent
}

type processorsStorage interface {
	Add(processor domain.ProcessorClientInfo)
	Get() []domain.ProcessorClientInfo
	IsBusy(processorID uint64) bool
	SetBusy(processorID uint64)
	SetFree(processorID uint64)
}

type metricsStorage interface {
	IncedentProcessed(incedent domain.Incedent, processor domain.IncedentProcessor)
	IncedentRejected(incedent domain.Incedent)
	PrintStatistics()
	ProcessInedent(incedent domain.Incedent, processor domain.IncedentProcessor)
	ReceivedIncedent(incedent domain.Incedent)
	RegisteredProcessor(processor domain.IncedentProcessor)
}

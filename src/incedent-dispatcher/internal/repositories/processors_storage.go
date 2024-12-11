package repositories

import (
	"slices"
	"sync"

	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/domain"
)

type ProcessorStorage struct {
	mu             sync.RWMutex
	processors     []domain.ProcessorClientInfo
	busyProcessors map[uint64]bool
}

func NewProcessorStorage() *ProcessorStorage {
	return &ProcessorStorage{
		processors:     make([]domain.ProcessorClientInfo, 0),
		busyProcessors: make(map[uint64]bool),
	}
}

func (ps *ProcessorStorage) Add(processor domain.ProcessorClientInfo) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.processors = append(ps.processors, processor)
}

func (ps *ProcessorStorage) Get() []domain.ProcessorClientInfo {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return slices.Clone(ps.processors)
}

func (ps *ProcessorStorage) IsBusy(processorID uint64) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return ps.busyProcessors[processorID]
}

func (ps *ProcessorStorage) SetBusy(processorID uint64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.busyProcessors[processorID] = true
}

func (ps *ProcessorStorage) SetFree(processorID uint64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.busyProcessors[processorID] = false
}

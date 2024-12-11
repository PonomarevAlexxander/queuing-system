package repositories

import (
	"sync"
	"time"

	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
	"github.com/benbjohnson/clock"
	"go.uber.org/zap"
)

type incedentStatus int

const (
	InBuffer incedentStatus = iota
	InProcessing
	Processed
	Rejected
)

type incedentInfo struct {
	status          incedentStatus
	processorID     uint64
	received        time.Time
	startProcessing time.Time
	endProcessing   time.Time
}

type processorInfo struct {
	regTime time.Time
	inWork  time.Duration
}

type producerStats struct {
	total                int
	rejected             int
	pRejected            float64
	timeInSystem         time.Duration
	timeInProcessing     time.Duration
	timeInBuffer         time.Duration
	dispTimeInBuffer     float64
	dispTimeInProcessing float64
}

type MetricsStorage struct {
	log *logger.Logger
	clk clock.Clock

	iMu        sync.Mutex
	incedents  map[domain.Priority]map[uint64]*incedentInfo
	pMu        sync.Mutex
	processors map[uint64]processorInfo
}

func NewMetricsStorage(log *logger.Logger, clk clock.Clock) *MetricsStorage {
	return &MetricsStorage{
		log:        log,
		clk:        clk,
		incedents:  make(map[domain.Priority]map[uint64]*incedentInfo),
		processors: make(map[uint64]processorInfo),
	}
}

func (ms *MetricsStorage) RegisteredProcessor(processor domain.IncedentProcessor) {
	ms.pMu.Lock()
	defer ms.pMu.Unlock()

	ms.processors[processor.Id] = processorInfo{
		regTime: ms.clk.Now(),
	}
}

func (ms *MetricsStorage) ReceivedIncedent(incedent domain.Incedent) {
	ms.iMu.Lock()
	defer ms.iMu.Unlock()

	info := ms.getIncedentInfo(incedent.Priority, incedent.Id)
	info.status = InBuffer
	info.received = incedent.CreationTime
}

func (ms *MetricsStorage) ProcessInedent(incedent domain.Incedent, processor domain.IncedentProcessor) {
	ms.iMu.Lock()
	defer ms.iMu.Unlock()

	info := ms.getIncedentInfo(incedent.Priority, incedent.Id)
	info.status = InProcessing
	info.startProcessing = ms.clk.Now()
	info.processorID = processor.Id
}

func (ms *MetricsStorage) IncedentProcessed(incedent domain.Incedent, processor domain.IncedentProcessor) {
	ms.iMu.Lock()
	defer ms.iMu.Unlock()

	info := ms.getIncedentInfo(incedent.Priority, incedent.Id)
	info.status = Processed
	info.endProcessing = ms.clk.Now()
}

func (ms *MetricsStorage) IncedentRejected(incedent domain.Incedent) {
	ms.iMu.Lock()
	defer ms.iMu.Unlock()

	info := ms.getIncedentInfo(incedent.Priority, incedent.Id)
	info.status = Rejected
}

func (ms *MetricsStorage) PrintStatistics() {
	ms.iMu.Lock()
	defer ms.iMu.Unlock()
	ms.pMu.Lock()
	defer ms.pMu.Unlock()
	currTime := ms.clk.Now()
	ms.log.Info("=== Statistics ===")
	for priority, incedents := range ms.incedents {
		stats := getIncedentStats(incedents, ms.processors)
		ms.log.Info("Producer statistics",
			zap.Any("priority", priority),
			zap.Int("total incedents", stats.total),
			zap.Int("number rejected", stats.rejected),
			zap.Float64("pRejected", stats.pRejected),
			zap.Stringer("timeInSystem", stats.timeInSystem),
			zap.Stringer("timeInProcessing", stats.timeInProcessing),
			zap.Stringer("timeInBuffer", stats.timeInBuffer),
			// zap.Float64("dispTimeInBuffer", stats.dispTimeInBuffer),
			// zap.Float64("dispTimeInProcessing", stats.dispTimeInProcessing),
		)
	}

	for id, info := range ms.processors {
		processorOn := currTime.Sub(info.regTime)
		ms.log.Info("Processors statistics",
			zap.Uint64("id", id),
			zap.Stringer("start time", info.regTime),
			zap.Stringer("end time", currTime),
			zap.Stringer("processorOn", processorOn),
			zap.Stringer("inWork", info.inWork),
			zap.Float64("utilityKoef", float64(info.inWork.Milliseconds())/float64(processorOn.Milliseconds())),
		)
	}
}

func (ms *MetricsStorage) getIncedentInfo(priority domain.Priority, id uint64) *incedentInfo {
	val, ok := ms.incedents[priority]
	if !ok {
		val = make(map[uint64]*incedentInfo)
		ms.incedents[priority] = val
	}

	currInfo, ok := val[id]
	if !ok {
		currInfo = &incedentInfo{}
		val[id] = currInfo
	}

	return currInfo
}

func getIncedentStats(incedents map[uint64]*incedentInfo, processors map[uint64]processorInfo) producerStats {
	stats := producerStats{}
	var (
		totalTimeInBuffer     time.Duration
		totalTimeInProcessing time.Duration
	)
	for _, incedent := range incedents {
		stats.total++
		if incedent.status != Processed {
			stats.rejected++
			continue
		}
		timeInBuffer := incedent.startProcessing.Sub(incedent.received)
		timeProcessing := incedent.endProcessing.Sub(incedent.startProcessing)
		totalTimeInBuffer += timeInBuffer
		totalTimeInProcessing += timeProcessing

		old := processors[incedent.processorID]
		old.inWork += timeProcessing
		processors[incedent.processorID] = old
	}
	stats.pRejected = float64(stats.rejected) / float64(stats.total)
	processed := stats.total - stats.rejected
	stats.timeInBuffer = time.Duration(totalTimeInBuffer.Milliseconds()/int64(processed)) * time.Millisecond
	stats.timeInProcessing = time.Duration(totalTimeInProcessing.Milliseconds()/int64(processed)) * time.Millisecond
	stats.timeInSystem = stats.timeInBuffer + stats.timeInProcessing
	return stats
}

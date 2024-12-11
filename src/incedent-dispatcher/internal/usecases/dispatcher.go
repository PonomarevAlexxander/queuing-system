package usecases

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
)

const (
	processorsChanCapacity = 10
	stopTimer              = 5 * time.Second
)

var (
	errIncedentEvicted    = errors.New("incedent was evicted by another one")
	errServiceUnavailable = errors.New("service is currentrly unavailable")
)

type incedentInfo struct {
	id       uint64
	priority domain.Priority
}

type IncedentDispatcher struct {
	log            *logger.Logger
	clk            clock.Clock
	bStorage       bufferStorage
	pStorage       processorsStorage
	metricsStorage metricsStorage

	stopped             chan struct{}
	mu                  sync.Mutex
	incedents           map[incedentInfo]chan error
	availableProcessors chan domain.ProcessorClientInfo
}

func NewIncedentDispatcher(
	log *logger.Logger,
	clk clock.Clock,
	bStorage bufferStorage,
	pStorage processorsStorage,
	metricsStorage metricsStorage,
) *IncedentDispatcher {
	return &IncedentDispatcher{
		log:            log,
		clk:            clk,
		bStorage:       bStorage,
		pStorage:       pStorage,
		metricsStorage: metricsStorage,
		stopped:        make(chan struct{}),
		incedents:      make(map[incedentInfo]chan error),
	}
}

func (ic *IncedentDispatcher) Run(ctx context.Context) error {
	return ic.runProcessing(ctx)
}

func (ic *IncedentDispatcher) Stop() {
}

func (ic *IncedentDispatcher) gracefulShutdown() {
	check := func() {
		ic.mu.Lock()
		defer ic.mu.Unlock()
		// wait all processed
		if len(ic.incedents) <= 0 {
			close(ic.stopped)
		}
	}
	timer := ic.clk.Timer(stopTimer)
	ic.log.Debug("Starting timer to shutdown dispatcher", zap.Stringer("timer", stopTimer))
	for {
		check()
		select {
		case <-timer.C:
			close(ic.stopped)
			ic.log.Info("Time exceeded, incedents wasn't handled completely")
			return
		case <-ic.stopped:
			ic.log.Info("All incedents processed, stopping dispatcher")
			return
		default:
		}
	}
}

func (ic *IncedentDispatcher) NewIncedent(ctx context.Context, incedent domain.Incedent) error {
	select {
	case <-ic.stopped:
		ic.log.Warn(
			"Rejected to process incedent, service terminating",
			zap.Stringer("incedent", incedent),
		)
		return errServiceUnavailable
	default:
	}

	ic.log.Info("New incedent received", zap.Stringer("incedent", incedent))
	ic.metricsStorage.ReceivedIncedent(incedent)

	wait := ic.newIncedent(incedent)
	if err := <-wait; err != nil {
		ic.metricsStorage.IncedentRejected(incedent)
		ic.log.Warn(
			"Incedent processed with error",
			zap.Stringer("incedent", incedent),
			zap.Error(err),
		)

		return err
	}
	ic.log.Info(
		"Incedent processed successfully",
		zap.Stringer("incedent", incedent),
	)

	return nil
}

func (ic *IncedentDispatcher) runProcessing(ctx context.Context) error {
	ic.availableProcessors = ic.fetchProcessors(ctx)
	var once sync.Once
	for {
		select {
		case <-ic.stopped:
			return nil
		case <-ctx.Done():
			once.Do(ic.gracefulShutdown)
		default:
		}

		packet := ic.bStorage.GetPacket()
		if len(packet) == 0 {
			continue
		}

		ic.log.Debug("Start to process packet", zap.Any("packet", packet))
		ic.processPacket(ctx, packet)
	}
}

func (ic *IncedentDispatcher) newIncedent(incedent domain.Incedent) chan error {
	waitChan := ic.createNewWaitChan(incedentInfo{
		id:       incedent.Id,
		priority: incedent.Priority,
	})

	if err := ic.bStorage.CheckAndPut(incedent); err == nil {
		return waitChan
	}

	// there is a chance to evict incedent in process
	evicted := ic.bStorage.EvictAndPut(incedent)
	ic.sendResult(
		incedentInfo{
			id:       evicted.Id,
			priority: evicted.Priority,
		},
		errIncedentEvicted,
	)

	return waitChan
}

func (ic *IncedentDispatcher) createNewWaitChan(info incedentInfo) chan error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ch := make(chan error, 1)
	ic.incedents[info] = ch

	return ch
}

func (ic *IncedentDispatcher) sendResult(info incedentInfo, result error) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ch, ok := ic.incedents[info]
	if !ok {
		ic.log.Error("Wait channel not found!")
		panic("wait channel not found") // should never happen
	}

	ch <- result
	close(ch)
	delete(ic.incedents, info)
}

func (ic *IncedentDispatcher) fetchProcessors(_ context.Context) chan domain.ProcessorClientInfo {
	ch := make(chan domain.ProcessorClientInfo, processorsChanCapacity)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(ch)
		wg.Done()

		for {
			select {
			case <-ic.stopped:
				return
			default:
			}

			processors := ic.pStorage.Get()
			for _, p := range processors {
				if ic.pStorage.IsBusy(p.Processor.Id) {
					continue
				}

				ic.pStorage.SetBusy(p.Processor.Id)
				ch <- p
			}
		}
	}()
	wg.Wait()

	return ch
}

// processPacket blocks until packet is processed
func (ic *IncedentDispatcher) processPacket(ctx context.Context, packet []domain.Incedent) {
	var eg errgroup.Group
	for _, incedent := range packet {
		processor := <-ic.availableProcessors
		eg.Go(func() error {
			ic.log.Debug("Processor is BUSY", zap.Stringer("processor", processor))
			ic.metricsStorage.ProcessInedent(incedent, processor.Processor)
			err := processor.Client.SendIncedent(ctx, incedent)
			ic.metricsStorage.IncedentProcessed(incedent, processor.Processor)
			ic.sendResult(incedentInfo{id: incedent.Id, priority: incedent.Priority}, err)
			err = ic.bStorage.DeleteIncedent(incedent)
			if err != nil {
				ic.log.Fatal("Buffer violation", zap.Error(err))
			}
			ic.log.Debug("Processor is FREE", zap.Stringer("processor", processor))
			ic.pStorage.SetFree(processor.Processor.Id)

			return nil
		})
	}
	eg.Wait()
}

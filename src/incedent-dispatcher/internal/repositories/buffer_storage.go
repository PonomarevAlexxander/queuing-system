package repositories

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher/internal/domain"
	"github.com/PonomarevAlexxander/queuing-system/utils/logger"
)

var (
	errBufferFull      = errors.New("buffer is full")
	errNothingToEvict  = errors.New("there are no incedents to evict")
	errElementNotFound = errors.New("element not found")
)

type BufferStorage struct {
	log         *logger.Logger
	mu          sync.Mutex
	maxCapacity int
	currentSize int
	buffer      map[domain.Priority][]domain.Incedent
}

func NewBufferStorage(log *logger.Logger, bufferCapacity uint64) *BufferStorage {
	return &BufferStorage{
		log:         log,
		maxCapacity: int(bufferCapacity),
		buffer:      make(map[domain.Priority][]domain.Incedent),
	}
}

func (bs *BufferStorage) CheckAndPut(incedent domain.Incedent) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	if bs.currentSize >= bs.maxCapacity {
		return errBufferFull
	}
	bs.putIncedent(incedent)

	return nil
}

func (bs *BufferStorage) EvictAndPut(incedent domain.Incedent) domain.Incedent {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	for priority := range bs.buffer {
		if priority < incedent.Priority {
			evicted, err := bs.evictOldest(priority)
			if err != nil {
				// try all lower priorities
				continue
			}
			bs.putIncedent(incedent)

			return evicted
		}
	}
	evicted, err := bs.evictOldest(incedent.Priority)
	if err != nil {
		return incedent
	}
	bs.putIncedent(incedent)

	return evicted
}

func (bs *BufferStorage) GetPacket() []domain.Incedent {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	var maxPriority domain.Priority
	for priority := range bs.buffer {
		if priority > maxPriority {
			maxPriority = priority
		}
	}

	return bs.ejectPacket(maxPriority)
}

func (bs *BufferStorage) DeleteIncedent(incedent domain.Incedent) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	err := bs.deleteIncedent(incedent)
	if err != nil {
		bs.currentSize--
	}

	return nil
}

func (bs *BufferStorage) evictOldest(priority domain.Priority) (domain.Incedent, error) {
	packet := bs.getPacket(priority)
	if len(packet) == 0 {
		return domain.Incedent{}, errNothingToEvict
	}
	// sort in descending order by creation time
	slices.SortFunc(packet, func(a, b domain.Incedent) int {
		return a.CreationTime.Compare(b.CreationTime) * -1
	})
	incedent := packet[len(packet)-1]
	packet = packet[:len(packet)-1]
	bs.buffer[priority] = packet
	bs.currentSize--

	return incedent, nil
}

func (bs *BufferStorage) putIncedent(incedent domain.Incedent) {
	// check slice exists
	_ = bs.getPacket(incedent.Priority)

	bs.buffer[incedent.Priority] = append(bs.buffer[incedent.Priority], incedent)
	bs.currentSize++
}

func (bs *BufferStorage) getPacket(priority domain.Priority) []domain.Incedent {
	packet, ok := bs.buffer[priority]
	if !ok {
		packet = make([]domain.Incedent, 0, bs.maxCapacity)
		bs.buffer[priority] = packet
	}

	return packet
}

func (bs *BufferStorage) ejectPacket(priority domain.Priority) []domain.Incedent {
	oldPacket := bs.getPacket(priority)
	bs.buffer[priority] = make([]domain.Incedent, 0, bs.maxCapacity)

	return oldPacket
}

func (bs *BufferStorage) deleteIncedent(incedent domain.Incedent) error {
	packet := bs.getPacket(incedent.Priority)
	for i, curr := range packet {
		if curr.Id == incedent.Id {
			packet = sliceDelete(packet, i)
			bs.buffer[incedent.Priority] = packet
			bs.currentSize--

			return nil
		}
	}

	return errElementNotFound
}

func sliceDelete[T any](slice []T, index int) []T {
	if index >= len(slice) {
		panic(fmt.Sprintf("cant delete from slice with len (%d) requested index (%d)", len(slice), index))
	}
	slice[index] = slice[len(slice)-1]

	return slice[:len(slice)-1]
}

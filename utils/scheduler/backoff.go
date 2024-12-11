package scheduler

import (
	"math"
	"time"
)

type LinearBackoff struct {
	startInterval time.Duration
}

func NewLinearBackoff(startInterval time.Duration) *LinearBackoff {
	return &LinearBackoff{
		startInterval: startInterval,
	}
}

func (l *LinearBackoff) NextInterval() time.Duration {
	return l.startInterval
}

type ExponentialBackoff struct {
	startInterval time.Duration
	counter       uint64
}

func NewExponentialBackoff(startInterval time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		startInterval: startInterval,
		counter:       0,
	}
}

func (e *ExponentialBackoff) NextInterval() time.Duration {
	defer func() { e.counter++ }()

	return time.Duration(math.Pow(float64(e.startInterval.Nanoseconds()), float64(e.counter))) * time.Nanosecond
}

package domain

import (
	"context"
	"fmt"
)

type IncedentProcessor struct {
	Id   uint64
	Host string
}

func (i IncedentProcessor) String() string {
	return fmt.Sprintf("Processor{%v, %v}", i.Id, i.Host)
}

type processorClient interface {
	SendIncedent(ctx context.Context, incedent Incedent) error
}

type ProcessorClientInfo struct {
	Processor IncedentProcessor
	Client    processorClient
}

func (i ProcessorClientInfo) String() string {
	return i.Processor.String()
}

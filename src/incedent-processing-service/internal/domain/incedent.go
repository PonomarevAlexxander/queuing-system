package domain

import (
	"fmt"
	"time"
)

type Priority uint64

type Incedent struct {
	Id           uint64
	CreationTime time.Time
	Priority     Priority
}

func (i Incedent) String() string {
	return fmt.Sprintf("Incedent{%v, %v, %v}", i.Id, i.CreationTime, i.Priority)
}

package config

import (
	"time"

	common_config "github.com/PonomarevAlexxander/queuing-system/utils/config"
)

type IncedentProducerConfig struct {
	common_config.CommonConfig `yaml:",inline"`
	InnerConfig                InnerConfig                `yaml:"incedent-producer" validate:"required"`
	DispatcherConfig           common_config.ClientConfig `yaml:"dispatcher" validate:"required"`
}

type InnerConfig struct {
	Interval string `yaml:"interval" validate:"required"`
}

func (ic InnerConfig) GetInterval() time.Duration {
	interval, err := time.ParseDuration(ic.Interval)
	if err != nil {
		panic(err)
	}

	return interval
}

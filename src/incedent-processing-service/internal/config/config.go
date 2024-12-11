package config

import (
	"strconv"
	"strings"
	"time"

	common_config "github.com/PonomarevAlexxander/queuing-system/utils/config"
)

type IncedentProcessorConfig struct {
	common_config.CommonConfig `yaml:",inline"`
	InnerConfig                InnerConfig                `yaml:"incedent-processor" validate:"required"`
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

func GetPort(host string) int {
	port, err := strconv.Atoi(strings.Split(host, ":")[1])
	if err != nil {
		panic(err)
	}

	return port
}

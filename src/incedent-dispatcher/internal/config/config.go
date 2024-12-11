package config

import (
	common_config "github.com/PonomarevAlexxander/queuing-system/utils/config"
)

type DispatcherConfig struct {
	common_config.CommonConfig `yaml:",inline"`
	InnerConfig                InnerConfig `yaml:"dispatcher" validate:"required"`
}

type InnerConfig struct {
	Port           int    `yaml:"port" validate:"required"`
	BufferCapacity uint64 `yaml:"buffer-capacity" validate:"required"`
}

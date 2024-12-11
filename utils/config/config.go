package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

type CommonConfig struct {
	LoggerConfig LoggerConfig `yaml:"logger" validate:"required"`
}

type LoggerConfig struct {
	Level      string   `yaml:"level" validate:"required,oneof='info' 'error' 'warn' 'debug'"`
	Out        []string `yaml:"out" validate:"required"` // can use stdout or paths to the file
	Type       string   `yaml:"type" validate:"required,oneof='console' 'json'"`
	Stacktrace bool     `yaml:"stacktrace" validate:"required"`
}

type ClientConfig struct {
	Host string `yaml:"host" validate:"required,hostname_port"`
}

var (
	validate *validator.Validate
	once     sync.Once
)

func ValidateConfig(config interface{}) error {
	once.Do(func() {
		validate = validator.New()
	})

	err := validate.Struct(config)
	if err != nil {
		return err
	}
	return nil
}

func ReadConfigFromYAML[T any](path string) (*T, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from %s, due to: %w", path, err)
	}
	var conf T
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from %s, due to: %w", path, err)
	}

	return &conf, nil
}

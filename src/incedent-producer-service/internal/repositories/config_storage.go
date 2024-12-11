package repositories

import (
	"sync"

	"github.com/PonomarevAlexxander/queuing-system/incedent-producer-service/internal/config"
)

type ConfigStorage struct {
	mu  sync.RWMutex
	cfg *config.IncedentProducerConfig
}

func NewConfigStorage() ConfigStorage {
	return ConfigStorage{}
}

func (c *ConfigStorage) SetConfig(cfg *config.IncedentProducerConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cfg = cfg
}

func (c *ConfigStorage) GetConfig() *config.IncedentProducerConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.cfg
}

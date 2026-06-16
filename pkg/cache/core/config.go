package core

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

type Engine string

const (
	MEMORY Engine = "memory"
)

type Config struct {
	Engine Engine       `mapstructure:"engine"`
	Memory MemoryConfig `mapstructure:"memory"`
}

func (c *Config) Validate() error {
	switch c.Engine {
	case MEMORY:
		return c.Memory.Validate()
	default:
		return fmt.Errorf("unknown caching engine: %s", c.Engine)
	}
}

func DefaultConfig() Config {
	return Config{
		Engine: MEMORY,
		Memory: MemoryConfig{
			MaxCost:     10_485_760,
			NumCounters: 10_000,
			BufferItems: 64,
			// TODO use a better default
			DefaultTTL: 5 * time.Second,
		},
	}
}

func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("cache", pflag.ExitOnError)

	cfg := DefaultConfig()
	fs.String("cache.engine", string(cfg.Engine), "Caching engine")
	fs.Int64("cache.memory.max_cost",
		cfg.Memory.MaxCost,
		"Max cost for memory cache",
	)
	fs.Int64("cache.memory.num_counters",
		cfg.Memory.NumCounters,
		"Number of counters for memory cache",
	)
	fs.Int64("cache.memory.buffer_items",
		cfg.Memory.BufferItems,
		"Number of buffer items for memory cache",
	)
	fs.Duration("cache.memory.default_ttl",
		cfg.Memory.DefaultTTL,
		"Default TTL for memory cache",
	)

	return fs
}

type MemoryConfig struct {
	MaxCost     int64         `mapstructure:"max_cost"`
	NumCounters int64         `mapstructure:"num_counters"`
	BufferItems int64         `mapstructure:"buffer_items"`
	DefaultTTL  time.Duration `mapstructure:"default_ttl"`
}

func (m *MemoryConfig) Validate() error {
	return nil
}

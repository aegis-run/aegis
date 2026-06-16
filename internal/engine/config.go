package engine

import "github.com/spf13/pflag"

type Config struct {
	MaxDepth                   int `mapstructure:"max_depth"`
	ExpressionConcurrencyLimit int `mapstructure:"expression_concurrency_limit"`
	TraversalConcurrencyLimit  int `mapstructure:"traversal_concurrency_limit"`
	PageSize                   int `mapstructure:"page_size"`
}

func (c *Config) Validate() error {
	return nil
}

func DefaultConfig() Config {
	return Config{
		MaxDepth:                   100,
		ExpressionConcurrencyLimit: 50,
		TraversalConcurrencyLimit:  50,
		PageSize:                   1000,
	}
}

func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("engine", pflag.ExitOnError)

	cfg := DefaultConfig()
	fs.Int("engine.max_depth", cfg.MaxDepth, "Check depth")
	fs.Int("engine.expression_concurrency_limit",
		cfg.ExpressionConcurrencyLimit,
		"Expression concurrency limit",
	)
	fs.Int("engine.traversal_concurrency_limit",
		cfg.TraversalConcurrencyLimit,
		"Traversal concurrency limit",
	)
	fs.Int("engine.page_size", cfg.PageSize, "Page size")

	return fs
}

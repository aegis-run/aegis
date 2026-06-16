package check

type Config struct {
	ExpressionConcurrencyLimit int
	TraversalConcurrencyLimit  int
	PageSize                   int32
}

func DefaultConfig() Config {
	return Config{
		ExpressionConcurrencyLimit: 50,
		TraversalConcurrencyLimit:  50,
		PageSize:                   1000,
	}
}

func (c Config) withDefaults() Config {
	defaults := DefaultConfig()

	if c.ExpressionConcurrencyLimit <= 0 {
		c.ExpressionConcurrencyLimit = defaults.ExpressionConcurrencyLimit
	}
	if c.TraversalConcurrencyLimit <= 0 {
		c.TraversalConcurrencyLimit = defaults.TraversalConcurrencyLimit
	}
	if c.PageSize <= 0 {
		c.PageSize = defaults.PageSize
	}

	return c
}

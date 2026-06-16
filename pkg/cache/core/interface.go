package core

import "time"

type Cache[V any] interface {
	Get(key string) (V, bool)
	Set(key string, value V, cost int64, ttl time.Duration) bool
	Wait()
	Close()
}

package retry

import (
	"math"
	"math/rand/v2"
	"time"
)

type BackoffFunc func(n int) time.Duration

func LinBackoff(delay time.Duration) BackoffFunc {
	return func(n int) time.Duration {
		if n < 1 {
			n = 1
		}

		return time.Duration(n) * delay
	}
}

func ExpBackoff(
	init time.Duration,
	mult, randFactor float64,
	maxInterval time.Duration,
) BackoffFunc {
	return func(n int) time.Duration {
		if n < 1 {
			n = 1
		}

		curr := float64(init) * math.Pow(mult, float64(n-1))
		curr = math.Min(curr, float64(maxInterval))

		delta := randFactor * curr
		// #nosec G404
		jitter := curr - delta + 2*delta*rand.Float64()

		return time.Duration(jitter)
	}
}

func DefaultExpBackoff() BackoffFunc {
	return ExpBackoff(100*time.Millisecond, 1.5, 0.5, 60*time.Second)
}

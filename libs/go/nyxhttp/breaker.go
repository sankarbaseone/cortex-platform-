package nyxhttp

import (
	"time"

	"github.com/sony/gobreaker/v2"
)

// NewBreaker returns a circuit breaker for one outbound dependency, tripped
// per the platform default (RFC-014 G.0): 5 consecutive failures -> open
// 30s -> half-open single probe. Wrap every outbound sync call in one.
func NewBreaker[T any](name string) *gobreaker.CircuitBreaker[T] {
	return gobreaker.NewCircuitBreaker[T](gobreaker.Settings{
		Name:        name,
		MaxRequests: 1,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})
}

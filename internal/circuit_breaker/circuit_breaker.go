package circuitbreaker

import (
	"sync"
	"time"
)

type CircuitBreaker struct {
	failureThreshold int
	failureCount     int
	lastFailure      time.Time
	timeout          time.Duration
}

var (
	checker = make(map[string]*CircuitBreaker)
	mu      sync.Mutex
)

func (cb *CircuitBreaker) CanRequest() bool {
	return true
}

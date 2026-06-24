package resilience

import (
	"errors"
	"shared/metrics"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type State string

const (
	StateClosed   State = "CLOSED"
	StateOpen     State = "OPEN"
	StateHalfOpen State = "HALF_OPEN"
)

type CircuitBreaker struct {
	serviceName      string
	state            State
	failureThreshold int
	failures         int
	successes        int
	lastError        time.Time
	timeout          time.Duration
	registry         *metrics.MetricsRegistry
	mu               sync.Mutex
}

func NewCircuitBreaker(serviceName string, failureThreshold int, timeout time.Duration, registry *metrics.MetricsRegistry) *CircuitBreaker {
	cb := &CircuitBreaker{
		serviceName:      serviceName,
		state:            StateClosed,
		failureThreshold: failureThreshold,
		timeout:          timeout,
		registry:         registry,
	}
	cb.updateMetric()
	return cb
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	switch cb.state {
	case StateClosed:
		cb.mu.Unlock()
		err := fn()
		cb.mu.Lock()
		defer cb.mu.Unlock()
		if err != nil {
			cb.failures++
			if cb.failures >= cb.failureThreshold {
				cb.transitionTo(StateOpen)
				cb.lastError = time.Now()
			}
			return err
		}
		cb.failures = 0
		return nil

	case StateOpen:
		if time.Since(cb.lastError) > cb.timeout {
			cb.transitionTo(StateHalfOpen)
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
		fallthrough

	case StateHalfOpen:
		cb.mu.Unlock()
		err := fn()
		cb.mu.Lock()
		defer cb.mu.Unlock()
		if err != nil {
			cb.transitionTo(StateOpen)
			cb.lastError = time.Now()
			return err
		}
		cb.transitionTo(StateClosed)
		cb.successes = 0
		cb.failures = 0
		return nil
	}

	cb.mu.Unlock()
	return nil
}

func (cb *CircuitBreaker) transitionTo(newState State) {
	cb.state = newState
	cb.updateMetric()
}

func (cb *CircuitBreaker) updateMetric() {
	if cb.registry != nil {
		// Reset all states to 0 first
		cb.registry.CircuitBreakerState.WithLabelValues(cb.serviceName, string(StateClosed)).Set(0)
		cb.registry.CircuitBreakerState.WithLabelValues(cb.serviceName, string(StateOpen)).Set(0)
		cb.registry.CircuitBreakerState.WithLabelValues(cb.serviceName, string(StateHalfOpen)).Set(0)

		// Set active state to 1
		cb.registry.CircuitBreakerState.WithLabelValues(cb.serviceName, string(cb.state)).Set(1)
	}
}

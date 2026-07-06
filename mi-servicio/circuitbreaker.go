package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// CircuitState representa los tres estados del circuit breaker: CLOSED (llamadas
// normales), OPEN (llamadas rechazadas sin intentar el RPC) y HALF_OPEN (se deja
// pasar una sola llamada de prueba para decidir si vuelve a CLOSED u OPEN).
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

var (
	circuitBreakerState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "circuit_breaker_state",
		Help: "Estado actual del circuit breaker por dependencia: 0=closed, 1=open, 2=half_open.",
	}, []string{"dependency"})
	circuitBreakerTransitionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "circuit_breaker_transitions_total",
		Help: "Transiciones de estado del circuit breaker por dependencia.",
	}, []string{"dependency", "to_state"})
	circuitBreakerShortCircuitsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "circuit_breaker_short_circuits_total",
		Help: "Llamadas rechazadas sin intentar el RPC porque el circuit breaker estaba abierto.",
	}, []string{"dependency"})
)

func init() {
	prometheus.MustRegister(
		circuitBreakerState,
		circuitBreakerTransitionsTotal,
		circuitBreakerShortCircuitsTotal,
	)
}

// ErrCircuitOpen se devuelve cuando una llamada se rechaza sin intentar el RPC
// porque el breaker de esa dependencia esta abierto (o ya hay una llamada de
// prueba en curso en half-open).
type ErrCircuitOpen struct {
	Dependency string
}

func (e *ErrCircuitOpen) Error() string {
	return fmt.Sprintf("circuit breaker abierto para %s", e.Dependency)
}

// CircuitBreaker es un circuit breaker independiente por dependencia: cuenta
// fallos consecutivos, abre tras un umbral, y tras un tiempo de espera deja
// pasar una unica llamada de prueba en half-open.
type CircuitBreaker struct {
	mu sync.Mutex

	name             string
	failureThreshold int
	openDuration     time.Duration

	state               CircuitState
	consecutiveFailures int
	openedAt            time.Time
	halfOpenInFlight    bool
}

func NewCircuitBreaker(name string, failureThreshold int, openDuration time.Duration) *CircuitBreaker {
	circuitBreakerState.WithLabelValues(name).Set(float64(StateClosed))
	return &CircuitBreaker{
		name:             name,
		failureThreshold: failureThreshold,
		openDuration:     openDuration,
		state:            StateClosed,
	}
}

// allow decide si una llamada puede intentarse. Si retorna false, el llamador
// debe fallar de inmediato sin intentar el RPC.
func (cb *CircuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		if time.Since(cb.openedAt) < cb.openDuration {
			return false
		}
		cb.transitionLocked(StateHalfOpen)
		cb.halfOpenInFlight = true
		return true

	case StateHalfOpen:
		if cb.halfOpenInFlight {
			return false
		}
		cb.halfOpenInFlight = true
		return true

	default:
		return false
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.halfOpenInFlight = false
	if cb.state != StateClosed {
		cb.transitionLocked(StateClosed)
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.halfOpenInFlight = false

	if cb.state == StateHalfOpen {
		cb.transitionLocked(StateOpen)
		return
	}

	cb.consecutiveFailures++
	if cb.state == StateClosed && cb.consecutiveFailures >= cb.failureThreshold {
		cb.transitionLocked(StateOpen)
	}
}

// transitionLocked asume que cb.mu ya esta tomado.
func (cb *CircuitBreaker) transitionLocked(to CircuitState) {
	from := cb.state
	cb.state = to
	if to == StateOpen {
		cb.openedAt = time.Now()
	}
	if to == StateClosed {
		cb.consecutiveFailures = 0
	}
	circuitBreakerState.WithLabelValues(cb.name).Set(float64(to))
	circuitBreakerTransitionsTotal.WithLabelValues(cb.name, to.String()).Inc()
	log.Printf("[CircuitBreaker] dependency=%s state=%s (desde %s)", cb.name, to.String(), from.String())
}

// Execute ejecuta fn si el breaker lo permite. Si el breaker esta abierto (o ya
// hay una llamada de prueba en half-open en curso), retorna *ErrCircuitOpen sin
// invocar fn. El resultado de fn actualiza el estado del breaker.
func Execute[T any](cb *CircuitBreaker, fn func() (T, error)) (T, error) {
	var zero T
	if !cb.allow() {
		circuitBreakerShortCircuitsTotal.WithLabelValues(cb.name).Inc()
		return zero, &ErrCircuitOpen{Dependency: cb.name}
	}

	result, err := fn()
	if err != nil {
		cb.recordFailure()
		return zero, err
	}

	cb.recordSuccess()
	return result, nil
}

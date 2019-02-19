package service

import (
	"github.com/kyokan/plasma/pkg/log"
	"github.com/sirupsen/logrus"
	"sync/atomic"
)

var breakerLogger = log.ForSubsystem("CircuitBreaker")

type CircuitBreaker interface {
	Tripped() bool
	Trip()
	Reset()
}

type CircuitBreakerImpl struct {
	tripped uint32

	logger *logrus.Entry
}

func NewCircuitBreaker(name string) CircuitBreaker {
	return &CircuitBreakerImpl{
		logger: breakerLogger.WithFields(logrus.Fields{
			"name": name,
		}),
	}
}

func (c *CircuitBreakerImpl) Tripped() bool {
	return atomic.LoadUint32(&c.tripped) == 1
}

func (c *CircuitBreakerImpl) Trip() {
	c.logger.Warn("circuit breaker tripped")
	atomic.StoreUint32(&c.tripped, 1)
}

func (c *CircuitBreakerImpl) Reset() {
	c.logger.Warn("circuit breaker reset")
	atomic.StoreUint32(&c.tripped, 0)
}

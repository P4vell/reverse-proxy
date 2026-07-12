package loadbalancer

import (
	"errors"
	"sync/atomic"

	"github.com/P4vell/reverse-proxy/internal/backend"
)

type LoadBalancer struct {
	backends       []*backend.Backend
	nextBackendIdx atomic.Int32
}

func NewLoadBalancer(backends []*backend.Backend) *LoadBalancer {
	return &LoadBalancer{
		backends: backends,
	}
}

func (lb *LoadBalancer) NextBackend() (*backend.Backend, error) {
	if len(lb.backends) == 0 {
		return nil, errors.New("no backends configured")
	}

	start := int(lb.nextBackendIdx.Add(1)-1) % len(lb.backends)

	for i := 0; i < len(lb.backends); i++ {
		idx := (start + i) % len(lb.backends)

		if lb.backends[idx].IsHealthy() {
			return lb.backends[idx], nil
		}
	}

	return nil, errors.New("no healthy backends found")
}

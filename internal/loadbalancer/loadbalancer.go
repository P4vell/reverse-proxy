package loadbalancer

import (
	"errors"
	"sync/atomic"

	"github.com/P4vell/reverse-proxy/internal/backend"
	"github.com/P4vell/reverse-proxy/internal/config"
)

type LoadBalancer struct {
	backends       []backend.Backend
	nextBackendIdx atomic.Int32
}

func NewLoadBalancer(servers []config.ServerConfig) *LoadBalancer {
	backends := make([]backend.Backend, 0, len(servers))

	for _, s := range servers {
		backends = append(backends, backend.Backend{
			Name:     s.Name,
			Protocol: s.Protocol,
			Host:     s.Host,
		})
	}

	return &LoadBalancer{
		backends: backends,
	}
}

func (lb *LoadBalancer) NextBackend() (backend.Backend, error) {
	if len(lb.backends) == 0 {
		return backend.Backend{}, errors.New("no healthy backends")
	}

	counter := lb.nextBackendIdx.Add(1)
	idx := int(counter-1) % len(lb.backends)

	return lb.backends[idx], nil
}

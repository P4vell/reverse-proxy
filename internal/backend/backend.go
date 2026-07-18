package backend

import (
	"sync/atomic"

	"github.com/P4vell/reverse-proxy/internal/config"
)

type Backend struct {
	Host     string
	Protocol string
	Name     string
	healthy  atomic.Bool
}

func LoadBackends(servers []config.ServerConfig) []*Backend {
	backends := make([]*Backend, 0, len(servers))

	for _, s := range servers {
		backends = append(backends, newBackend(s.Name, s.Protocol, s.Host))
	}

	return backends
}

func newBackend(name, protocol, host string) *Backend {
	b := &Backend{
		Name:     name,
		Protocol: protocol,
		Host:     host,
	}

	b.healthy.Store(true)

	return b
}

func (b *Backend) IsHealthy() bool {
	return b.healthy.Load()
}

func (b *Backend) MarkHealthy() {
	b.healthy.Store(true)
}

func (b *Backend) MarkUnhealthy() {
	b.healthy.Store(false)
}

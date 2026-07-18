package proxy

import "github.com/P4vell/reverse-proxy/internal/backend"

type BackendSelector interface {
	NextBackend() (*backend.Backend, error)
	GetNumBackends() int
}

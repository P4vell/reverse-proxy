package healthchecker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/P4vell/reverse-proxy/internal/backend"
	"github.com/P4vell/reverse-proxy/internal/config"
)

type HealthChecker struct {
	backends []*backend.Backend
	interval time.Duration
	client   *http.Client
}

func New(cfg config.HealthChecker, backends []*backend.Backend) *HealthChecker {
	return &HealthChecker{
		backends: backends,
		interval: cfg.Interval * time.Second,
		client: &http.Client{
			Timeout: cfg.Timeout * time.Second,
		},
	}
}

func (hc *HealthChecker) Start(ctx context.Context) {
	// Synchronous health check on startup
	for _, b := range hc.backends {
		hc.healthCheck(b)
	}

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, b := range hc.backends {
				go hc.healthCheck(b)
			}
		}
	}
}

func (hc *HealthChecker) healthCheck(b *backend.Backend) {
	url := fmt.Sprintf("%s://%s/health", b.Protocol, b.Host)
	res, err := hc.client.Get(url)
	if err != nil {
		b.MarkUnhealthy()
		return
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		b.MarkHealthy()
	} else {
		b.MarkUnhealthy()
	}
}

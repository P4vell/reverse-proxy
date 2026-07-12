package main

import (
	"context"
	"log"

	"github.com/P4vell/reverse-proxy/internal/backend"
	"github.com/P4vell/reverse-proxy/internal/config"
	"github.com/P4vell/reverse-proxy/internal/healthcheck"
	"github.com/P4vell/reverse-proxy/internal/loadbalancer"
	"github.com/P4vell/reverse-proxy/internal/proxy"
	"github.com/P4vell/reverse-proxy/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	ctx := context.Background()

	backends := backend.LoadBackends(cfg.Servers)

	loadBalancer := loadbalancer.NewLoadBalancer(backends)

	healthChecker := healthcheck.NewHealthChecker(cfg.HealthCheck, backends)
	go healthChecker.Start(ctx)

	proxyHandler := proxy.NewProxy(loadBalancer)

	httpServer := server.NewServer(cfg, proxyHandler)
	err = httpServer.ListenAndServe()
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	backends := backend.LoadBackends(cfg.Servers)
	loadBalancer := loadbalancer.NewLoadBalancer(backends)

	healthChecker := healthcheck.NewHealthChecker(cfg.HealthCheck, backends)
	go healthChecker.Start(ctx)

	proxyHandler := proxy.NewProxy(loadBalancer)
	httpServer := server.NewServer(cfg, proxyHandler)

	go runHTTPServer(httpServer)

	<-ctx.Done()
	shutdownServer(httpServer)
}

func runHTTPServer(srv *http.Server) {
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func shutdownServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
	}

	log.Println("server stopped")
}

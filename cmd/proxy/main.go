package main

import (
	"log"

	"github.com/P4vell/reverse-proxy/internal/config"
	"github.com/P4vell/reverse-proxy/internal/loadbalancer"
	"github.com/P4vell/reverse-proxy/internal/proxy"
	"github.com/P4vell/reverse-proxy/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	loadBalancer := loadbalancer.NewLoadBalancer(cfg.Servers)
	proxyHandler := proxy.NewProxy(loadBalancer)
	httpServer := server.NewServer(cfg, proxyHandler)
	err = httpServer.ListenAndServe()
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

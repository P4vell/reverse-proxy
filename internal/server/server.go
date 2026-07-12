package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/P4vell/reverse-proxy/internal/config"
)

func NewServer(cfg config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

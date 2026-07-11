package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/P4vell/reverse-proxy/internal/config"
)

type backend struct {
	host     string
	protocol string
	name     string
}

type Proxy struct {
	client         *http.Client
	backends       []backend
	nextBackendIdx atomic.Int32
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	proxy := NewProxy(cfg)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		targetBackend, err := proxy.nextBackend()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		outReq := proxy.getOutboundRequest(req, targetBackend)

		res, err := proxy.client.Do(outReq)
		if err != nil {
			fmt.Println(err)
			w.Write([]byte("Failed to forward request"))
			return
		}

		defer res.Body.Close()

		err = proxy.copyResponse(res, w)
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	http.ListenAndServe(":8080", nil)
}

func NewProxy(cfg config.Config) *Proxy {
	backends := make([]backend, 0, len(cfg.Servers))

	for _, s := range cfg.Servers {
		backends = append(backends, backend{
			name:     s.Name,
			protocol: s.Protocol,
			host:     s.Host,
		})
	}

	return &Proxy{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		backends: backends,
	}
}

func (p *Proxy) nextBackend() (backend, error) {
	if len(p.backends) == 0 {
		return backend{}, errors.New("no healthy backends")
	}

	counter := p.nextBackendIdx.Add(1)
	idx := int(counter-1) % len(p.backends)

	return p.backends[idx], nil
}

func (p *Proxy) getOutboundRequest(req *http.Request, targetBackend backend) *http.Request {
	headers := make(http.Header)
	maps.Copy(headers, req.Header)

	return &http.Request{
		Method: req.Method,
		Body:   req.Body,
		Header: headers,
		URL: &url.URL{
			Scheme:   targetBackend.protocol,
			Host:     targetBackend.host,
			Path:     req.URL.Path,
			RawQuery: req.URL.RawQuery,
		},
	}
}

func (p *Proxy) copyResponse(res *http.Response, w http.ResponseWriter) error {
	for key, values := range res.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(res.StatusCode)

	_, err := io.Copy(w, res.Body)
	if err != nil {
		return err
	}

	return nil
}

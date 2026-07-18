package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"time"

	"github.com/P4vell/reverse-proxy/internal/backend"
)

type Proxy struct {
	client   *http.Client
	selector BackendSelector
}

func New(selector BackendSelector) *Proxy {
	return &Proxy{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		selector: selector,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targetBackend, err := p.selector.NextBackend()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	res, err := p.forward(r, targetBackend)
	if err != nil {
		res, err = p.forwardWithRetry(r, targetBackend)
		if err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
	}
	defer res.Body.Close()

	err = copyResponse(res, w)
	if err != nil {
		log.Println(err)
	}
}

func (p *Proxy) forward(req *http.Request, b *backend.Backend) (*http.Response, error) {
	log.Printf("%s %s -> %s\n", req.Method, req.URL.Path, b.Name)
	outReq := getOutboundRequest(req, b)
	res, err := p.client.Do(outReq)
	if err != nil {
		return nil, fmt.Errorf("backend %s failed: %w", b.Name, err)
	}

	return res, nil
}

func (p *Proxy) forwardWithRetry(req *http.Request, failedBackend *backend.Backend) (*http.Response, error) {
	if !isRetryableMethod(req.Method) {
		return nil, errors.New("request method is not retryable")
	}

	failedBackend.MarkUnhealthy()

	for range p.selector.BackendsNum() {
		nextBackend, err := p.selector.NextBackend()
		if err != nil {
			continue
		}

		res, err := p.forward(req, nextBackend)
		if err != nil {
			continue
		}

		return res, nil
	}

	return nil, errors.New("all retry attempts failed")
}

func isRetryableMethod(method string) bool {
	switch method {
	case http.MethodGet,
		http.MethodHead,
		http.MethodOptions:
		return true
	default:
		return false
	}
}

func getOutboundRequest(req *http.Request, targetBackend *backend.Backend) *http.Request {
	headers := make(http.Header)
	maps.Copy(headers, req.Header)

	return &http.Request{
		Method: req.Method,
		Body:   req.Body,
		Header: headers,
		URL: &url.URL{
			Scheme:   targetBackend.Protocol,
			Host:     targetBackend.Host,
			Path:     req.URL.Path,
			RawQuery: req.URL.RawQuery,
		},
	}
}

func copyResponse(res *http.Response, w http.ResponseWriter) error {
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

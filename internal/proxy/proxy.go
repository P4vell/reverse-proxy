package proxy

import (
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

func NewProxy(selector BackendSelector) *Proxy {
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

	log.Printf("%s %s -> %s\n", r.Method, r.URL.Path, targetBackend.Name)

	outReq := getOutboundRequest(r, targetBackend)

	res, err := p.client.Do(outReq)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Failed to forward request"))
		return
	}

	defer res.Body.Close()

	err = p.copyResponse(res, w)
	if err != nil {
		log.Println(err)
		return
	}
}

func getOutboundRequest(req *http.Request, targetBackend backend.Backend) *http.Request {
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

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	name string
	port int
}

func main() {
	namePtr := flag.String("name", "", "name of the server")
	portPtr := flag.Int("port", 0, "port used by the server")

	flag.Parse()

	server, err := NewServer(*namePtr, *portPtr)
	if err != nil {
		log.Fatalf("failed to create new server: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res := fmt.Sprintf("Hello from server %s", server.name)
		w.Write([]byte(res))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	addr := fmt.Sprintf(":%d", server.port)
	fmt.Printf("Server %s listening on %s", server.name, addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func NewServer(name string, port int) (*Server, error) {
	if name == "" || port == 0 {
		return nil, errors.New("missing name or port")
	}

	return &Server{
		name: name,
		port: port,
	}, nil
}

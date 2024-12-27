package service

import (
	"context"
	"net/http"
	"time"

	"github.com/a-h/templ"
)

type Server struct {
	Mux *http.ServeMux
	srv *http.Server
}

func NewServer(addr string) *Server {
	m := http.NewServeMux()
	m.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	m.Handle("/", templ.Handler(Home()))
	return &Server{
		Mux: m,
		srv: &http.Server{Addr: addr, Handler: m},
	}
}

// Wait on the returned channel to block until the server stops.
func (s *Server) ListenAndServe() <-chan error {
	errors := make(chan error, 1)
	go func() {
		errors <- s.srv.ListenAndServe()
		close(errors)
	}()

	return errors
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.srv.Shutdown(ctx)
}

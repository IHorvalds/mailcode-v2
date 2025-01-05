package service

import (
	"context"
	"log"
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
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			errorHandler(w, r, http.StatusNotFound)
			return
		}

		templ.Handler(Home()).ServeHTTP(w, r)
	})
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

func (s *Server) GetAddress() string {
	return s.srv.Addr
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)

	if status == http.StatusNotFound {
		log.Printf("404 status: tried '%s'", r.URL.Path)
		ErrorView(status, "Page not found").Render(r.Context(), w)
		return
	}
	ErrorView(status, "").Render(r.Context(), w)
}

package auth

import (
	"net/http"

	"github.com/a-h/templ"
)

func RegisterAuthHomeHandler(s *http.ServeMux) {
	s.Handle("GET /auth/", templ.Handler(AuthHome()))
}

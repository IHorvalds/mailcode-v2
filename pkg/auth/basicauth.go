package auth

import (
	"net/http"
)

type BasicAuthCredentials struct {
	User     string
	Password string
}

// NewBasicAuth returns a new BasicAuthCredentials with the given user and password.
func NewBasicAuth(user, password string) *BasicAuthCredentials {
	return &BasicAuthCredentials{
		User:     user,
		Password: password,
	}
}

// Repository for the BasicAuth
type BasicAuthRepository interface {
	CreateBasicAuthCredsTables() error
	SaveBasicAuthCreds(ba *BasicAuthCredentials) error
	DeleteBasicAuthCreds(user string) error
	GetBasicAuthCreds(user string) (*BasicAuthCredentials, error)
	ListBasicAuthCreds(filter string) ([]BasicAuthCredentials, error)
}

func RegisterBasicAuthHandlers(s *http.ServeMux, repo BasicAuthRepository) {
	s.Handle("GET /auth/basic/all", getBasicAuthCreds(repo))
	s.Handle("POST /auth/basic/new", saveBasicAuthCreds(repo))
	s.Handle("POST /auth/basic/{user}", updateBasicAuthCreds(repo))
	s.Handle("DELETE /auth/basic/{user}", deleteBasicAuthCreds(repo))
}

func getBasicAuthCreds(repo BasicAuthRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bas, err := repo.ListBasicAuthCreds("")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		BasicAuthCredsList(bas).Render(r.Context(), w)
	}
}

func saveBasicAuthCreds(repo BasicAuthRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ba := &BasicAuthCredentials{
			User:     r.FormValue("username"),
			Password: r.FormValue("password"),
		}

		isUpdate := false
		if u, err := repo.GetBasicAuthCreds(ba.User); u != nil && err == nil {
			isUpdate = true
		}
		err := repo.SaveBasicAuthCreds(ba)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if !isUpdate {
			basicAuthTableCell(*ba).Render(r.Context(), w)
		}
	}
}

func updateBasicAuthCreds(repo BasicAuthRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.PathValue("user")
		if user == "" {
			http.Error(w, "missing user query parameter", http.StatusBadRequest)
			return
		}
		ba := &BasicAuthCredentials{
			User:     r.FormValue("username"),
			Password: r.FormValue("password"),
		}
		err := repo.SaveBasicAuthCreds(ba)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteBasicAuthCreds(repo BasicAuthRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.PathValue("user")
		if user == "" {
			http.Error(w, "missing user query parameter", http.StatusBadRequest)
			return
		}
		err := repo.DeleteBasicAuthCreds(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

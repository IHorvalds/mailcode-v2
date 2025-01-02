package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

type OAuth2Credentials struct {
	User         string
	Provider     string
	AccessToken  string
	RefreshToken string
	ExpiryDate   time.Time
}

func NewOAuth2(user, provider, accessToken, refreshToken string, expiryDate time.Time) *OAuth2Credentials {
	return &OAuth2Credentials{
		User:         user,
		Provider:     provider,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiryDate:   expiryDate,
	}
}

func FromGothUser(u *goth.User) *OAuth2Credentials {
	return &OAuth2Credentials{
		User:         u.UserID,
		Provider:     u.Provider,
		AccessToken:  u.AccessToken,
		RefreshToken: u.RefreshToken,
		ExpiryDate:   u.ExpiresAt,
	}
}

func ToGothUser(o *OAuth2Credentials) *goth.User {
	return &goth.User{
		UserID:       o.User,
		Provider:     o.Provider,
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
		ExpiresAt:    o.ExpiryDate,
	}
}

type OAuth2Repository interface {
	CreateOAuth2CredsTables() error
	SaveOAuth2Creds(ba *OAuth2Credentials) error
	DeleteOAuth2Creds(user string) error
	GetOAuth2Creds(user string) (*OAuth2Credentials, error)
	ListCreds(filter string) ([]OAuth2Credentials, error)
}

func RegisterOAuth2Handlers(s *http.ServeMux, repo OAuth2Repository) {
	s.Handle("GET /auth/oauth2/all", getAllOAuth2Creds(repo))
	s.Handle("DELETE /auth/oauth2/{user}", deleteOAuth2Creds(repo))
	s.Handle("GET /auth/oauth2/{provider}/login", authenticate(repo))
	s.Handle("GET /auth/oauth2/{provider}/callback", oauth2Callback(repo))
}

func getAllOAuth2Creds(repo OAuth2Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		creds, err := repo.ListCreds("")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		OAuth2CredsList(creds).Render(r.Context(), w)
	}
}

func deleteOAuth2Creds(repo OAuth2Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if user == "" {
			http.Error(w, "missing user", http.StatusBadRequest)
			return
		}
		err := repo.DeleteOAuth2Creds(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type contextKey string

const providerKey contextKey = "provider"

func authenticate(repo OAuth2Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		if provider == "" {
			http.Error(w, "missing provider", http.StatusBadRequest)
			return
		}

		if provider == "" {
			http.Error(w, "missing provider", http.StatusBadRequest)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), providerKey, provider))

		if u, err := gothic.CompleteUserAuth(w, r); err == nil {
			if err := repo.SaveOAuth2Creds(FromGothUser(&u)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/auth/oauth2/all", http.StatusSeeOther)
			return
		}

		gothic.BeginAuthHandler(w, r)
	}
}

func oauth2Callback(repo OAuth2Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		if provider == "" {
			http.Error(w, "missing provider", http.StatusBadRequest)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), providerKey, provider))

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err := repo.SaveOAuth2Creds(FromGothUser(&user)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		http.Redirect(w, r, "/auth/oauth2/all", http.StatusSeeOther)
	}
}

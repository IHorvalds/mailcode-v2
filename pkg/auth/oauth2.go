package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
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
		User:         u.Email,
		Provider:     u.Provider,
		AccessToken:  u.AccessToken,
		RefreshToken: u.RefreshToken,
		ExpiryDate:   u.ExpiresAt,
	}
}

func ToGothUser(o *OAuth2Credentials) *goth.User {
	return &goth.User{
		Email:        o.User,
		Provider:     o.Provider,
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
		ExpiresAt:    o.ExpiryDate,
	}
}

// These should be set by the build script
// -ldflags "-X ${PACKAGE}/pkg/auth.GOOGLE_CLIENT_ID=..."
// See the README
var GOOGLE_CLIENT_ID string
var GOOGLE_CLIENT_SECRET string

func init() {
	cookieStore := sessions.NewCookieStore()
	cookieStore.Options.HttpOnly = true
	gothic.Store = cookieStore
}

type OAuth2Repository interface {
	CreateOAuth2CredsTables() error
	SaveOAuth2Creds(ba *OAuth2Credentials) error
	DeleteOAuth2Creds(user string) error
	GetOAuth2Creds(user string) (*OAuth2Credentials, error)
	ListCreds(filter string) ([]OAuth2Credentials, error)
}

func RegisterOAuth2Handlers(s *http.ServeMux, serverAddr string, repo OAuth2Repository) {
	s.Handle("GET /auth/oauth2/all", getAllOAuth2Creds(repo))
	s.Handle("DELETE /auth/oauth2/{user}", deleteOAuth2Creds(repo))
	s.Handle("GET /auth/oauth2/{provider}/login", authenticate(repo))
	s.Handle("GET /auth/oauth2/{provider}/callback", oauth2Callback(repo))

	goth.UseProviders(
		google.New(GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, fmt.Sprintf("http://%s/auth/oauth2/google/callback", serverAddr), "email", "https://mail.google.com/"),
	)
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

func authenticate(OAuth2Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := r.PathValue("provider")
		if provider == "" {
			http.Error(w, "missing provider", http.StatusBadRequest)
			return
		}

		if provider == "" {
			http.Error(w, "missing provider", http.StatusBadRequest)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

		gothic.BeginAuthHandler(w, r)
	}
}

func oauth2Callback(repo OAuth2Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := r.PathValue("provider")
		if provider == "" {
			http.Error(w, "missing provider", http.StatusBadRequest)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else if err := repo.SaveOAuth2Creds(FromGothUser(&user)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err == nil {
			http.Redirect(w, r, "/auth/oauth2/all", http.StatusSeeOther)
		}
	}
}

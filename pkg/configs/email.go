package configs

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

//go:generate stringer -type Encryption
type Encryption int

const (
	EncryptionNone Encryption = iota
	EncryptionStartTLS
	EncryptionSSL
)

var EncryptionValueFromName = func() map[string]Encryption {
	m := make(map[string]Encryption)
	for i := EncryptionNone; i <= EncryptionSSL; i++ {
		m[i.String()] = i
	}
	return m
}()

type Email struct {
	User       string
	IMAPServer string
	Port       uint16
	Encryption Encryption
	Inbox      string
}

// Default Email configuration
// SSL encryption (rather than StartTLS)
// INBOX as the default Inbox.
// Check the email provider's IMAP settings
func NewEmail() *Email {
	return &Email{
		Port:       993,
		Encryption: EncryptionSSL,
		Inbox:      "INBOX",
	}
}

// CRUD operations for the Email configuration
// using db as the storage layer.

type EmailRepository interface {
	CreateEmailTables() error
	SaveEmail(e *Email) error
	ListEmails(filter string) ([]Email, error)
	GetEmail(user string) (*Email, error)
	DeleteEmail(user string) error
}

var ErrNoResult = errors.New("no result")

// HTTP Handlers for the Email configuration
func Register(s *http.ServeMux, repo EmailRepository) {
	s.HandleFunc("GET /emails", HandleGetEmails(repo))
	s.HandleFunc("GET /emails/new", HandleNewEmail())
	s.HandleFunc("POST /emails/new", HandlePostEmail(repo))
	s.HandleFunc("GET /emails/{user}", HandleGetEmail(repo))
	s.HandleFunc("PUT /emails/{user}", HandlePutEmail(repo))
	s.HandleFunc("DELETE /emails/{user}", HandleDeleteEmail(repo))
}

func HandleGetEmails(repo EmailRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emails, err := repo.ListEmails(r.URL.Query().Get("q"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		component := EmailList(emails)
		component.Render(r.Context(), w)
	}
}

func HandleGetEmail(repo EmailRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.PathValue("user")
		email, err := repo.GetEmail(user)
		if err != nil {
			if errors.Is(err, ErrNoResult) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			ErrorView(err.Error()).Render(r.Context(), w)
			return
		}

		EmailForm(email).Render(r.Context(), w)
	}
}

func HandleNewEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		EmailForm(nil).Render(r.Context(), w)
	}
}

func HandlePostEmail(repo EmailRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		p, err := strconv.ParseInt(r.FormValue("port"), 10, 16)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			ErrorView(err.Error()).Render(r.Context(), w)
			return
		}

		enc, ok := EncryptionValueFromName[r.FormValue("encryption")]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			ErrorView("invalid encryption value").Render(r.Context(), w)
			return
		}

		email := &Email{
			User:       r.FormValue("user"),
			IMAPServer: r.FormValue("server"),
			Port:       uint16(p),
			Encryption: enc,
			Inbox:      r.FormValue("inbox"),
		}

		if err := repo.SaveEmail(email); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ErrorView(err.Error()).Render(r.Context(), w)
			return
		}

		w.WriteHeader(http.StatusOK)
		EmailForm(email).Render(r.Context(), w)
	}
}

func HandlePutEmail(repo EmailRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.PathValue("user")
		email := &Email{}
		if err := json.NewDecoder(r.Body).Decode(email); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			ErrorView(err.Error()).Render(r.Context(), w)
			return
		}

		if _, err := repo.GetEmail(user); err != nil {
			if errors.Is(err, ErrNoResult) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			ErrorView(err.Error()).Render(r.Context(), w)
			return
		}

		if err := repo.SaveEmail(email); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ErrorView(err.Error()).Render(r.Context(), w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleDeleteEmail(repo EmailRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.PathValue("user")
		if err := repo.DeleteEmail(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ErrorView(err.Error()).Render(r.Context(), w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

package configs

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Encryption int

// go:generate stringer -type=Encryption
const (
	EncryptionNone Encryption = iota
	EncryptionStartTLS
	EncryptionSSL
)

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

var NoResult = errors.New("no result")

// HTTP Handlers for the Email configuration
func Register(s *http.ServeMux, repo EmailRepository) {
	s.HandleFunc("GET /emails", HandleGetEmails(repo))
	s.HandleFunc("POST /emails", HandlePostEmail(repo))
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

		// TODO: Use templ to create templates and return the appropriate one here
		json.NewEncoder(w).Encode(emails)
	}
}

func HandleGetEmail(repo EmailRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.PathValue("user")
		email, err := repo.GetEmail(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO: Use templ to create templates and return the appropriate one here
		json.NewEncoder(w).Encode(email)
	}
}

func HandlePostEmail(repo EmailRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := &Email{}
		if err := json.NewDecoder(r.Body).Decode(email); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := repo.SaveEmail(email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandlePutEmail(repo EmailRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.PathValue("user")
		email := &Email{}
		if err := json.NewDecoder(r.Body).Decode(email); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := repo.GetEmail(user); err != nil {
			if errors.Is(err, NoResult) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := repo.SaveEmail(email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleDeleteEmail(repo EmailRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.PathValue("user")
		if err := repo.DeleteEmail(user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

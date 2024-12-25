package persistence

import (
	"database/sql"
	"errors"

	"github.com/IHorvalds/mailcode-v2/pkg/auth"
	"github.com/IHorvalds/mailcode-v2/pkg/configs"
	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(path string) (*Repository, error) {
	db, err := sql.Open("sqlite3", path)

	if err != nil {
		return nil, err
	}

	repo := &Repository{db: db}

	repo.CreateEmailTables()
	repo.CreateBasicAuthCredsTables()
	repo.CreateOAuth2CredsTables()

	return repo, nil
}

func (r *Repository) Shutdown() {
	r.db.Close()
}

// EmailRepository methods
func (r *Repository) CreateEmailTables() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS emails (
		user TEXT PRIMARY KEY,
		imap_server TEXT NOT NULL,
		port INTEGER NOT NULL,
		encryption INTEGER NOT NULL CHECK (encryption >= 0 AND encryption <= 2) DEFAULT 2,
		inbox TEXT NOT NULL DEFAULT "INBOX"
	);`)
	return err
}

func (r *Repository) SaveEmail(email *configs.Email) error {
	_, err := r.db.Exec(`INSERT INTO emails (user, imap_server, port, encryption, inbox) VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(user) DO UPDATE SET imap_server = excluded.imap_server, port = excluded.port, encryption = excluded.encryption, inbox = excluded.inbox;
	`, email.User, email.IMAPServer, email.Port, email.Encryption, email.Inbox)
	return err
}

func (r *Repository) GetEmail(user string) (*configs.Email, error) {
	email := &configs.Email{}
	err := r.db.QueryRow("SELECT user, imap_server, port, encryption, inbox FROM emails WHERE user = ?", user).Scan(&email.User, &email.IMAPServer, &email.Port, &email.Encryption, &email.Inbox)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("email not found")
		}
		return nil, err
	}
	return email, nil
}

func (r *Repository) DeleteEmail(user string) error {
	_, err := r.db.Exec("DELETE FROM emails WHERE user = ?", user)
	return err
}

func (r *Repository) ListEmails(query string) ([]configs.Email, error) {
	rows, err := r.db.Query("SELECT user, imap_server, port, encryption, inbox FROM emails WHERE user LIKE ?", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []configs.Email
	for rows.Next() {
		var email configs.Email
		if err := rows.Scan(&email.User, &email.IMAPServer, &email.Port, &email.Encryption, &email.Inbox); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

// BasicAuthCredentialsRepository methods
func (r *Repository) CreateBasicAuthCredsTables() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS basic_auth (
		username TEXT PRIMARY KEY,
		password TEXT NOT NULL,
		FORGIGN KEY (username) REFERENCES emails(user) ON DELETE CASCADE
	);`)
	return err
}

func (r *Repository) SaveBasicAuthCredentials(creds *auth.BasicAuthCredentials) error {
	_, err := r.db.Exec(`INSERT INTO basic_auth (username, password) VALUES (?, ?)
	ON CONFLICT(username) DO UPDATE SET password = excluded.password; 
	`, creds.User, creds.Password)
	return err
}

func (r *Repository) GetBasicAuthCredentials(username string) (*auth.BasicAuthCredentials, error) {
	creds := &auth.BasicAuthCredentials{}
	err := r.db.QueryRow("SELECT username, password FROM basic_auth WHERE username = ?", username).Scan(&creds.User, &creds.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("credentials not found")
		}
		return nil, err
	}
	return creds, nil
}

func (r *Repository) DeleteBasicAuthCredentials(user string) error {
	_, err := r.db.Exec("DELETE FROM basic_auth WHERE username = ?", user)
	return err
}

// OAuth2CredentialsRepository methods
func (r *Repository) CreateOAuth2CredsTables() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS oauth2_tokens (
		username TEXT PRIMARY KEY,
		access_token TEXT NOT NULL,
		refresh_token TEXT,
		expiry_date DATETIME,
		FORGIGN KEY (username) REFERENCES emails(user) ON DELETE CASCADE
	);`)
	return err
}

func (r *Repository) SaveOAuth2Token(token *auth.OAuth2Credentials) error {
	_, err := r.db.Exec(`INSERT INTO oauth2_tokens (username, access_token, refresh_token, expiry_date) VALUES (?, ?, ?, ?)
	ON CONFLICT(username) DO UPDATE SET access_token = excluded.access_token, refresh_token = excluded.refresh_token, expiry_date = excluded.expiry_date; 
	`, token.User, token.AccessToken, token.RefreshToken, token.ExpiryDate)
	return err
}

func (r *Repository) GetOAuth2Creds(username string) (*auth.OAuth2Credentials, error) {
	creds := &auth.OAuth2Credentials{}
	err := r.db.QueryRow("SELECT username, access_token, refresh_token, expiry_date FROM oauth2_tokens WHERE username = ?", username).Scan(&creds.User, &creds.AccessToken, &creds.RefreshToken, &creds.ExpiryDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("credentials not found")
		}
		return nil, err
	}
	return creds, nil
}

func (r *Repository) DeleteOAuth2Creds(user string) error {
	_, err := r.db.Exec("DELETE FROM oauth2_tokens WHERE username = ?", user)
	return err
}

package persistence

import (
	"database/sql"
	"errors"

	"github.com/IHorvalds/mailcode-v2/pkg/auth"
)

// OAuth2CredentialsRepository methods
func (r *Repository) CreateOAuth2CredsTables() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS oauth2_tokens (
		username TEXT PRIMARY KEY CHECK (username <> ''),
		provider TEXT CHECK (provider <> ''),
		access_token TEXT NOT NULL,
		refresh_token TEXT,
		expiry_date DATETIME,
		FORGIGN KEY username REFERENCES emails(user) ON DELETE CASCADE
	);`)
	return err
}

func (r *Repository) SaveOAuth2Creds(token *auth.OAuth2Credentials) error {
	_, err := r.db.Exec(`INSERT INTO oauth2_tokens (username, provider, access_token, refresh_token, expiry_date) VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(username) DO UPDATE SET provider = excluded.provider, access_token = excluded.access_token, refresh_token = excluded.refresh_token, expiry_date = excluded.expiry_date; 
	`, token.User, token.Provider, token.AccessToken, token.RefreshToken, token.ExpiryDate)
	return err
}

func (r *Repository) GetOAuth2Creds(username string) (*auth.OAuth2Credentials, error) {
	creds := &auth.OAuth2Credentials{}
	err := r.db.QueryRow("SELECT username, provider, access_token, refresh_token, expiry_date FROM oauth2_tokens WHERE username = ?", username).Scan(&creds.User, &creds.Provider, &creds.AccessToken, &creds.RefreshToken, &creds.ExpiryDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("credentials not found")
		}
		return nil, err
	}
	return creds, nil
}

func (r *Repository) ListCreds(filter string) ([]auth.OAuth2Credentials, error) {

	args := []interface{}{}
	query := "SELECT username, provider, access_token, refresh_token, expiry_date FROM oauth2_tokens"
	if filter != "" {
		query += " WHERE username LIKE ?"
		args = append(args, filter)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []auth.OAuth2Credentials
	for rows.Next() {
		cred := auth.OAuth2Credentials{}
		err := rows.Scan(&cred.User, &cred.Provider, &cred.AccessToken, &cred.RefreshToken, &cred.ExpiryDate)
		if err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, nil
}

func (r *Repository) DeleteOAuth2Creds(user string) error {
	_, err := r.db.Exec("DELETE FROM oauth2_tokens WHERE username = ?", user)
	return err
}

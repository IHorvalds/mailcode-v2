package persistence

import (
	"database/sql"
	"errors"

	"github.com/IHorvalds/mailcode-v2/pkg/auth"
)

func (r *Repository) CreateBasicAuthCredsTables() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS basic_auth (
		username TEXT PRIMARY KEY,
		password TEXT NOT NULL,
		FORGIGN KEY username REFERENCES emails(user) ON DELETE CASCADE
	);`)
	return err
}

func (r *Repository) SaveBasicAuthCreds(creds *auth.BasicAuthCredentials) error {
	_, err := r.db.Exec(`INSERT INTO basic_auth (username, password) VALUES (?, ?)
	ON CONFLICT(username) DO UPDATE SET password = excluded.password; 
	`, creds.User, creds.Password)
	return err
}

func (r *Repository) GetBasicAuthCreds(username string) (*auth.BasicAuthCredentials, error) {
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

func (r *Repository) ListBasicAuthCreds(filter string) ([]auth.BasicAuthCredentials, error) {

	args := []interface{}{}
	query := "SELECT username, password FROM basic_auth"
	if filter != "" {
		query += " WHERE username LIKE ?"
		args = append(args, filter)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []auth.BasicAuthCredentials
	for rows.Next() {
		cred := auth.BasicAuthCredentials{}
		err := rows.Scan(&cred.User, &cred.Password)
		if err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, nil
}

func (r *Repository) DeleteBasicAuthCreds(user string) error {
	_, err := r.db.Exec("DELETE FROM basic_auth WHERE username = ?", user)
	return err
}

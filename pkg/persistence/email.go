package persistence

import (
	"database/sql"
	"errors"

	"github.com/IHorvalds/mailcode-v2/pkg/configs"
)

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

package mailwatcher

import (
	"container/list"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	conn *sql.DB
}

type Mailbox struct {
	Email    string
	Password string
	Server   string
	Port     int32
	UseSSL   bool
}

func DefaultRepoPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Failed to get current working directory")
	}
	return path.Join(cwd, "emails.db")
}

func OpenRepository(dbPath string) (Repository, error) {
	con, err := sql.Open("sqlite3", dbPath)
	repo := Repository{}
	if err != nil {
		return repo, err
	}

	if con == nil {
		return repo, errors.New("failed to connect to db, null pointer")
	}

	mailboxesTableSql := `
	CREATE TABLE IF NOT EXISTS mailboxes (
		email TEXT PRIMARY KEY,
		password TEXT NOT NULL,
		server TEXT NOT NULL,
		port INTEGER NOT NULL,
		usessl BOOLEAN DEFAULT TRUE
	);`

	_, err = con.Exec(mailboxesTableSql)
	if err != nil {
		return repo, err
	}

	repo.conn = con
	return repo, nil
}

func (rep *Repository) Close() error {
	return rep.conn.Close()
}

func (rep *Repository) GetAllMailboxes() (*list.List, error) {
	var getMailbox = "SELECT email, password, server, port, usessl FROM mailboxes;"
	rows, err := rep.conn.Query(getMailbox)
	if err != nil {
		return list.New(), err
	}

	defer rows.Close()

	mailboxes := list.New()
	for rows.Next() {
		m := Mailbox{}
		err := rows.Scan(&(m.Email), &(m.Password), &(m.Server), &(m.Port), &(m.UseSSL))
		if err != nil {
			return list.New(), err
		}

		mailboxes.PushBack(&m)
	}

	return mailboxes, nil
}

func (rep *Repository) AddMailbox(m *Mailbox) error {
	var insertMailbox = `INSERT INTO mailboxes VALUES 
	(:email, :password, :server, :port, :useSSL);`
	_, err := rep.conn.Exec(insertMailbox,
		sql.Named("email", m.Email),
		sql.Named("password", m.Password),
		sql.Named("server", m.Server),
		sql.Named("port", m.Port),
		sql.Named("useSSL", m.UseSSL))
	return err
}

func (rep *Repository) RemoveMailbox(email string) error {
	var deleteMailbox = `DELETE FROM mailboxes WHERE
	email=:email;`
	_, err := rep.conn.Exec(deleteMailbox, email)
	return err
}

func (rep *Repository) GetMailbox(email string) (Mailbox, error) {
	var getMailbox = "SELECT email, password, server, port, usessl FROM mailboxes WHERE email=:email;"
	row := rep.conn.QueryRow(getMailbox, sql.Named("email", email))

	m := Mailbox{}
	err := row.Scan(&(m.Email), &(m.Password), &(m.Server), &(m.Port), &(m.UseSSL))
	if err != nil {
		return Mailbox{}, err
	}
	return m, nil
}

func (mb Mailbox) ToString() string {
	protocol := "imap"
	if mb.UseSSL {
		protocol = "imaps"
	}
	return fmt.Sprintf("email: %s\npassword: %s\nserver: %s://%s:%d\n\n", mb.Email, mb.Password, protocol, mb.Server, mb.Port)
}

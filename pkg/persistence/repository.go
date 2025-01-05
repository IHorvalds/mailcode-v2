package persistence

import (
	"database/sql"

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

	if err := repo.CreateEmailTables(); err != nil {
		return nil, err
	}

	if err := repo.CreateBasicAuthCredsTables(); err != nil {
		return nil, err
	}
	if err := repo.CreateOAuth2CredsTables(); err != nil {
		return nil, err
	}

	if err := repo.CreateSubjectTable(); err != nil {
		return nil, err
	}

	if err := repo.CreateExtractorTable(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) Shutdown() {
	r.db.Close()
}

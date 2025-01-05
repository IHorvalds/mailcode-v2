package persistence

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/IHorvalds/mailcode-v2/pkg/configs"
)

func (r *Repository) CreateSubjectTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS subjects (
		subject TEXT PRIMARY KEY CHECK (subject <> ''),
	);`)
	return err
}

func (r *Repository) GetSubjects(filter string) ([]configs.Subject, error) {
	query := "SELECT subject FROM subjects"
	args := []interface{}{}
	if filter != "" {
		query += " WHERE subject LIKE ?"
		args = append(args, filter)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subjects := []configs.Subject{}
	for rows.Next() {
		var sub configs.Subject
		if err := rows.Scan(&sub); err != nil {
			return nil, err
		}
		subjects = append(subjects, sub)
	}
	return subjects, nil
}

func (r *Repository) SaveSubjects(s []configs.Subject) error {
	if len(s) == 0 {
		return nil
	}

	placeholders := []string{}
	for range len(s) {
		placeholders = append(placeholders, "(?)")
	}
	_, err := r.db.Exec(fmt.Sprintf(`INSERT INTO subjects (subject) VALUES %s
	ON CONFLICT(subject) DO NOTHING;`, strings.Join(placeholders, ", ")), s)
	return err
}

func (r *Repository) DeleteSubject(s configs.Subject) error {
	_, err := r.db.Exec("DELETE FROM subjects WHERE subject = ?", s)
	return err
}

func (r *Repository) CreateExtractorTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS extractors (
		id BIGSERIAL PRIMARY KEY,
		expression TEXT UNIQUE CHECK (expression <> ''),
		capture TEXT CHECK (capture <> '')
	);`)
	return err
}

func (r *Repository) GetExtractors() ([]configs.Extractor, error) {
	query := "SELECT id, expression, capture FROM extractors"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	e := []configs.Extractor{}
	for rows.Next() {
		var id int
		var expr string
		var cap string
		if err := rows.Scan(&id, &expr, &cap); err != nil {
			return nil, err
		}
		re, err := regexp.Compile(expr)
		if err != nil {
			log.Printf("Invalid regular expression `%s`. Deleting extractor with ID `%d`", expr, id)
			r.DeleteExtractor(configs.Extractor{ID: id})
			continue
		}
		e = append(e, configs.Extractor{
			ID:      id,
			Expr:    *re,
			Capture: cap})
	}

	return e, nil
}

func (r *Repository) AddExtractor(e configs.Extractor) error {
	_, err := r.db.Exec(`INSERT INTO extractors (expression, capture) VALUES (?, ?)
	ON CONFLICT(expression) DO UPDATE SET capture = excluded.capture;`)
	return err
}

func (r *Repository) DeleteExtractor(e configs.Extractor) error {
	if e.ID != 0 {
		_, err := r.db.Exec("DELETE FROM extractors WHERE ID = ?", e.ID)
		return err
	}

	_, err := r.db.Exec("DELETE FROM extractors WHERE expression = ?", e.Expr.String())
	return err
}

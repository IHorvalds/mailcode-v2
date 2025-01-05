package configs

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
)

type Subject string

type Extractor struct {
	ID      int
	Expr    regexp.Regexp
	Capture string
}

type SubjectRepository interface {
	CreateSubjectTable() error
	GetSubjects(filter string) ([]Subject, error)
	SaveSubjects(s []Subject) error
	DeleteSubject(s Subject) error
}

type ExtractorRepository interface {
	CreateExtractorTable() error
	GetExtractors() ([]Extractor, error)
	AddExtractor(e Extractor) error
	DeleteExtractor(e Extractor) error
}

type ExtractorSubjectRepository interface {
	SubjectRepository
	ExtractorRepository
}

func RegisterExtractorHandlers(s *http.ServeMux, r ExtractorSubjectRepository) {
	s.Handle("GET /extractors", handleGetExtractorsAndSubjects(r))
	s.Handle("POST /extractors", handleNewExtractor(r))
	s.Handle("DELETE /extractors/{id}", handleDeleteExtractor(r))

	s.Handle("POST /subjects", handleSaveSubjects(r))
	s.Handle("DELETE /subjects/{subject}", handleDeleteSubject(r))
}

func handleGetExtractorsAndSubjects(repo ExtractorSubjectRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		extractors, err := repo.GetExtractors()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		subjects, err := repo.GetSubjects("")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ExtractorsAndSubjects(extractors, subjects).Render(r.Context(), w)
	}
}

func handleNewExtractor(repo ExtractorRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var extractor Extractor
		if err := json.NewDecoder(r.Body).Decode(&extractor); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := repo.AddExtractor(extractor); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleDeleteExtractor(repo ExtractorRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("id")

		id, err := strconv.Atoi(sid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if err := repo.DeleteExtractor(Extractor{ID: id}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func handleSaveSubjects(repo SubjectRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var subject Subject
		if err := json.NewDecoder(r.Body).Decode(&subject); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := repo.Save(&subject); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(subject)
	}
}

func handleDeleteSubject(repo SubjectRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sub := r.PathValue("subject")

		if err := repo.DeleteSubject(Subject(sub)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

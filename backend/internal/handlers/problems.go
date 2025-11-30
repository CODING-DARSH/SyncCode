package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"syncode/internal/models"
)


func ListProblems(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT id, title, slug, description, input_format, output_format, constraints, difficulty, created_at, updated_at
			FROM problems
			ORDER BY id
		`)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var problems []models.Problem
		for rows.Next() {
			var p models.Problem
			if err := rows.Scan(
				&p.ID, &p.Title, &p.Slug, &p.Description,
				&p.InputFormat, &p.OutputFormat, &p.Constraints,
				&p.Difficulty, &p.CreatedAt, &p.UpdatedAt,
			); err != nil {
				http.Error(w, "scan error", http.StatusInternalServerError)
				return
			}
			problems = append(problems, p)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(problems)
	}
}
func GetProblem(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var p models.Problem
		err = db.QueryRow(`
			SELECT id, title, slug, description, input_format, output_format, constraints, difficulty, created_at, updated_at
			FROM problems
			WHERE id = $1
		`, id).Scan(
			&p.ID, &p.Title, &p.Slug, &p.Description,
			&p.InputFormat, &p.OutputFormat, &p.Constraints,
			&p.Difficulty, &p.CreatedAt, &p.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "problem not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}
}


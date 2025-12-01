package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"syncode/internal/models"
)

type CreateSubmissionRequest struct {
	UserID     int    `json:"user_id"`
	ProblemID  int    `json:"problem_id"`
	LanguageID int    `json:"language_id"`
	Code       string `json:"code"`
}

func CreateSubmission(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateSubmissionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		var id int
		err := db.QueryRow(`
			INSERT INTO submissions (user_id, problem_id, language_id, code, status)
			VALUES ($1, $2, $3, $4, 'PENDING')
			RETURNING id
		`, req.UserID, req.ProblemID, req.LanguageID, req.Code).Scan(&id)

		if err != nil {
			http.Error(w, "db insert error", http.StatusInternalServerError)
			return
		}

		resp := map[string]any{
			"id":     id,
			"status": "PENDING",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}


func GetSubmission(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var s models.Submission
		err = db.QueryRow(`
			SELECT id, user_id, problem_id, language_id, code, status, exec_time, memory, output, error, failing_input, created_at
			FROM submissions
			WHERE id = $1
		`, id).Scan(
			&s.ID, &s.UserID, &s.ProblemID, &s.LanguageID, &s.Code,
			&s.Status, &s.ExecTime, &s.Memory, &s.Output, &s.Error,
			&s.FailingInput, &s.CreatedAt,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	}
}

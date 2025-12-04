package models

import "time"

type Submission struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	ProblemID    int       `json:"problem_id"`
	LanguageID   int       `json:"language_id"`
	Code         string    `json:"code"`
	Status       string    `json:"status"`
	ExecTime     float64   `json:"exec_time"`
	Memory       float64   `json:"memory"`
	Output       string    `json:"output"`
	Error        string    `json:"error"`
	FailingInput string    `json:"failing_input"`
	CreatedAt    time.Time `json:"created_at"`
}

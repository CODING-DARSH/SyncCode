package main

import (
	"fmt"
	"log"
	"net/http"

	"syncode/internal/database"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	_ = db

	fmt.Println("✅ Connected to PostgreSQL")
	fmt.Println("API server running on :8082")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Fatal(http.ListenAndServe(":8082", nil))
}

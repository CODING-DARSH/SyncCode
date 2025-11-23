package database

import (
    "database/sql"
    "fmt"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func Connect() (*sql.DB, error) {
    connStr := "postgres://postgres:Darsh@localhost:5432/SyncCode?sslmode=disable"

    db, err := sql.Open("pgx", connStr)
    if err != nil {
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        return nil, err
    }

    fmt.Println("✅ Connected to PostgreSQL")
    return db, nil
}

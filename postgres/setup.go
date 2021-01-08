package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const dialect = "postgres"

// Open returns a connection to the database if the credentials are valid.
func Open(connectionString string) (*sqlx.DB, error) {
	db, err := sqlx.Connect(dialect, connectionString)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %w", err)
	}
	return db, nil
}

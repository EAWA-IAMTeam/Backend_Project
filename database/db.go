package database

import (
	"backend_project/internal/config"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	env := config.LoadConfig()

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		env.DbHost, env.DbUser, env.DbPassword, env.DbName, env.DbPort, env.DbSSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening db: %v", err)
	}
	return db, nil
}

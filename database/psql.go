package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/lijuuu/EmployeeManagement/config"
)

func InitDB(cfg *config.Config) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS employees (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			position VARCHAR(255) NOT NULL,
			salary INTEGER NOT NULL,
			hired_date DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

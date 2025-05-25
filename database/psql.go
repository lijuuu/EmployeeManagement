package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func NewPostgresConn(ctx context.Context, connString string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

package tools

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func Create(ctx context.Context) *pgx.Conn {
	connStr := "postgres://postgres:postgres@localhost:5432/avito_shop"
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		// TODO logger
	}
	return conn
}

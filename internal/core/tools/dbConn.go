package tools

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func Create(ctx context.Context, logger *zap.Logger) *pgx.Conn {
	connStr := "postgres://postgres:postgres@localhost:5432/avito_shop"
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		logger.Error(
			"failed to connect to database",
			zap.Error(err),
			zap.String("conn_string", connStr),
		)
	}
	return conn
}

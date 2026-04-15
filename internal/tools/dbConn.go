package tools

import (
	"avito-shop/internal/config"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func Create(ctx context.Context, logger *zap.Logger) *pgx.Conn {
	connStr := fmt.Sprintf(
		"%v://%v:%v@%v:%v/%v",
		config.App.Storage.Connection.Driver,
		config.App.Storage.Connection.User,
		config.App.Storage.Connection.Password,
		config.App.Storage.Connection.Host,
		config.App.Storage.Connection.Port,
		config.App.Storage.Connection.Database,
	)
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

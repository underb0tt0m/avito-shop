package tools

import (
	"avito-shop/internal/config"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePool(ctx context.Context) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"%v://%v:%v@%v:%v/%v",
		config.App.Storage.Connection.Driver,
		config.App.Storage.Connection.User,
		config.App.Storage.Connection.Password,
		config.App.Storage.Connection.Host,
		config.App.Storage.Connection.Port,
		config.App.Storage.Connection.Database,
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

package postgres

import (
	"avito-shop/internal/logging"
	"avito-shop/internal/storage"
	"context"

	"github.com/jackc/pgx/v5"
)

type storageAuth struct {
	Conn   *pgx.Conn
	Logger logging.Logger
}

func NewStorageAuth(conn *pgx.Conn, logger logging.Logger) storage.Auth {
	return storageAuth{
		Conn:   conn,
		Logger: logger,
	}
}

func (s storageAuth) GetHashedUserPassword(ctx context.Context, username string) ([]byte, bool, error) {

	return []byte{}, true, nil
}

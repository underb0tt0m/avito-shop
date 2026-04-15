package postgres

import (
	"avito-shop/internal/storage"
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type storageAuth struct {
	Conn   *pgx.Conn
	Logger *zap.Logger
}

func NewStorageAuth(conn *pgx.Conn, logger *zap.Logger) storage.Auth {
	return storageAuth{
		Conn:   conn,
		Logger: logger,
	}
}

func (s storageAuth) GetHashedUserPassword(ctx context.Context, username string) ([]byte, bool, error) {

	return []byte{}, true, nil
}

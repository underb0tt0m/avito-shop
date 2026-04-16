package postgres

import (
	"avito-shop/internal/domain"
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

func (s storageAuth) GetHashedUserPassword(ctx context.Context, username string) ([]byte, error) {
	stmt := `
SELECT
	password_hash
FROM
	users
WHERE
    name = $1;
;`
	row := s.Conn.QueryRow(ctx, stmt, username)

	var hashedPassword []byte
	if err := row.Scan(&hashedPassword); err != nil {
		return nil, err
	}
	return hashedPassword, nil
}

func (s storageAuth) CreateUser(ctx context.Context, user domain.HashedUserData) ([]byte, error) {
	stmt := `
INSERT INTO users (name, password_hash)
VALUES ($1, $2)
;`
	if _, err := s.Conn.Exec(ctx, stmt, user.Name, user.Password); err != nil {
		return nil, err
	}
	return user.Password, nil
}

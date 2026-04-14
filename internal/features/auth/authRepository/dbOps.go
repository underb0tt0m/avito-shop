package authRepository

import (
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type StorageImpl struct {
	Conn   *pgx.Conn
	Logger *zap.Logger
}

func (s StorageImpl) GetHashedUserPassword(username string, password []byte) ([]byte, bool, error) {

	return password, true, nil
}

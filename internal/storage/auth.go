package storage

import (
	"avito-shop/internal/domain"
	"context"
)

type Auth interface {
	GetHashedUserPassword(ctx context.Context, username string) ([]byte, error)
	CreateUser(ctx context.Context, user domain.HashedUserData) ([]byte, error)
}

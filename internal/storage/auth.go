package storage

import (
	"context"

	"avito-shop/internal/domain"
)

//go:generate mockgen -source=auth.go -destination=../mocks/storage_auth.go -package=mocks -mock_names=Auth=StorageAuth
type Auth interface {
	GetHashedUserPassword(ctx context.Context, username string) ([]byte, error)
	CreateUser(ctx context.Context, user domain.HashedUserData) ([]byte, error)
}

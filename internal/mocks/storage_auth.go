package mocks

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/storage"
	"context"
)

type mockStorageAuth struct {
	GetHashedUserPasswordFunc func(ctx context.Context, username string) ([]byte, error)
	CreateUserFunc            func(ctx context.Context, user domain.HashedUserData) ([]byte, error)
}

func NewStorageAuth(
	GetHashedUserPasswordFunc func(ctx context.Context, username string) ([]byte, error),
	CreateUserFunc func(ctx context.Context, user domain.HashedUserData) ([]byte, error),
) storage.Auth {
	return mockStorageAuth{
		GetHashedUserPasswordFunc: GetHashedUserPasswordFunc,
		CreateUserFunc:            CreateUserFunc,
	}
}

func (s mockStorageAuth) GetHashedUserPassword(ctx context.Context, username string) ([]byte, error) {
	if s.GetHashedUserPasswordFunc != nil {
		return s.GetHashedUserPasswordFunc(ctx, username)
	}
	return []byte("test"), nil
}

func (s mockStorageAuth) CreateUser(ctx context.Context, user domain.HashedUserData) ([]byte, error) {
	if s.CreateUserFunc != nil {
		return s.CreateUserFunc(ctx, user)
	}
	return user.Password, nil
}

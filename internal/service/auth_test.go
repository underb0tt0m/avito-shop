package service

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/mocks"
	"avito-shop/internal/storage"
	"avito-shop/internal/tools"
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestAuth(t *testing.T) {
	tests := []struct {
		name            string
		data            dto.AuthRequest
		mockStorage     storage.Auth
		tokenMaker      tools.TokenMaker
		hasher          tools.Hasher
		expected        dto.AuthResponse
		wantErr         bool
		wantSpecificErr error
	}{
		{
			"success_new_user_created",
			dto.AuthRequest{Name: "test", Password: "test"},
			mocks.NewStorageAuth(nil, nil),
			mocks.NewToken(nil, nil, nil),
			mocks.NewHasher(nil, nil),
			dto.AuthResponse{Token: "test"},
			false,
			nil,
		},

		{
			"success_existing_user_authenticated",
			dto.AuthRequest{Name: "test", Password: "test"},
			mocks.NewStorageAuth(
				func(ctx context.Context, username string) ([]byte, error) {
					return []byte{}, pgx.ErrNoRows
				},
				nil,
			),
			mocks.NewToken(nil, nil, nil),
			mocks.NewHasher(nil, nil),
			dto.AuthResponse{Token: "test"},
			false,
			nil,
		},

		{
			"error_wrong_password",
			dto.AuthRequest{Name: "test", Password: "test"},
			mocks.NewStorageAuth(nil, nil),
			mocks.NewToken(nil, nil, nil),
			mocks.NewHasher(
				nil,
				func(hashedPassword []byte, password []byte) error {
					return errors.New("test")
				},
			),
			dto.AuthResponse{},
			true,
			domain.ErrUnauthorized,
		},

		{
			"error_database_unavailable_on_get",
			dto.AuthRequest{Name: "test", Password: "test"},
			mocks.NewStorageAuth(
				func(ctx context.Context, username string) ([]byte, error) {
					return []byte{}, errors.New("test")
				},
				nil,
			),
			mocks.NewToken(nil, nil, nil),
			mocks.NewHasher(nil, nil),
			dto.AuthResponse{},
			true,
			nil,
		},

		{
			"error_token_creation_failed",
			dto.AuthRequest{Name: "test", Password: "test"},
			mocks.NewStorageAuth(nil, nil),
			mocks.NewToken(
				func(data any) (string, error) {
					return "", errors.New("test")
				},
				nil,
				nil,
			),
			mocks.NewHasher(nil, nil),
			dto.AuthResponse{},
			true,
			nil,
		},

		{
			"error_failed_to_hash_password",
			dto.AuthRequest{Name: "test", Password: "test"},
			mocks.NewStorageAuth(nil, nil),
			mocks.NewToken(nil, nil, nil),
			mocks.NewHasher(
				func(data string, logger logging.Logger) ([]byte, error) {
					return []byte{}, domain.ErrInternalServerError
				},
				nil,
			),
			dto.AuthResponse{},
			true,
			domain.ErrInternalServerError,
		},

		{
			"error_failed_to_create_user",
			dto.AuthRequest{Name: "test", Password: "test"},
			mocks.NewStorageAuth(
				func(ctx context.Context, username string) ([]byte, error) {
					return nil, pgx.ErrNoRows
				},
				func(ctx context.Context, user domain.HashedUserData) ([]byte, error) {
					return nil, errors.New("test")
				},
			),
			mocks.NewToken(nil, nil, nil),
			mocks.NewHasher(nil, nil),
			dto.AuthResponse{},
			true,
			nil,
		},
	}

	if err := config.Init("../../cmd/config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := mocks.NewLogger(nil)

	for _, test := range tests {
		TestAuthService := NewAuth(
			test.mockStorage,
			logger,
			test.tokenMaker,
			test.hasher,
		)
		ctx, cancel := context.WithTimeout(context.Background(), config.App.Storage.QueryTimeout)

		result, err := TestAuthService.Auth(ctx, test.data)
		cancel()

		if err != nil {
			if !test.wantErr {
				t.Fatalf("Test %v, Auth() unexpected error: %v", test.name, err)
			}
			if test.wantSpecificErr != nil && !errors.Is(err, test.wantSpecificErr) {
				t.Errorf("Test %v, Auth() = %+v, want %+v", test.name, err, test.wantSpecificErr)
			}
		}

		if result != test.expected {
			t.Errorf("Test %v, Auth() = %+v, want %+v", test.name, result, test.expected)
		} else {
			t.Logf("Test %v, Auth() success: %+v", test.name, result)
		}
	}

}

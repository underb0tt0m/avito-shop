package service

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/storage"
	"avito-shop/internal/tools"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -source=auth.go -destination=../mocks/service_auth.go -package=mocks -mock_names=Auth=MockServiceAuth
type Auth interface {
	Auth(ctx context.Context, data dto.AuthRequest) (dto.AuthResponse, error)
}

type auth struct {
	Storage    storage.Auth
	Logger     logging.Logger
	TokenMaker tools.TokenMaker
	Hasher     tools.Hasher
}

func NewAuth(s storage.Auth, l logging.Logger, t tools.TokenMaker, h tools.Hasher) Auth {
	return auth{
		Storage:    s,
		Logger:     l,
		TokenMaker: t,
		Hasher:     h,
	}
}

func (s auth) Auth(ctx context.Context, data dto.AuthRequest) (dto.AuthResponse, error) {
	hashedPassword, err := s.Hasher.Hash(data.Password, s.Logger)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	hashedUser := domain.HashedUserData{
		Name:     data.Name,
		Password: hashedPassword,
	}

	DBHashedPassword, err := s.Storage.GetHashedUserPassword(ctx, hashedUser.Name)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		DBHashedPassword, err = s.Storage.CreateUser(ctx, hashedUser)
		if err != nil {
			return dto.AuthResponse{}, fmt.Errorf(
				"failed to create new user: %v",
				err,
			)
		}
		s.Logger.Info("create new user")
	case err != nil:
		return dto.AuthResponse{}, fmt.Errorf(
			"failed to get user password from Storage: %v",
			err,
		)

	}

	if err = s.Hasher.CompareHashAndPassword(
		DBHashedPassword,
		[]byte(data.Password),
	); err != nil {
		s.Logger.Warn(
			fmt.Sprintf(
				"wrong password: %v",
				hashedUser.Name,
			),
			domain.ErrUnauthorized,
		)

		return dto.AuthResponse{}, domain.ErrUnauthorized
	}

	userClaims := domain.DefaultUser{UserName: hashedUser.Name}
	token, err := s.TokenMaker.CreateToken(userClaims)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return dto.AuthResponse{Token: token}, nil
}

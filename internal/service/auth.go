package service

import (
	"context"
	"errors"
	"fmt"

	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/hasher"
	"avito-shop/internal/jwtmanager"
	"avito-shop/internal/logging"
	"avito-shop/internal/prometheus_metrics"
	"avito-shop/internal/storage"

	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -source=auth.go -destination=../mocks/service_auth.go -package=mocks -mock_names=Auth=ServiceAuth
type Auth interface {
	Auth(ctx context.Context, data dto.AuthRequest) (dto.AuthResponse, error)
}

type auth struct {
	Storage    storage.Auth
	Logger     logging.Logger
	TokenMaker jwtmanager.TokenMaker
	Hasher     hasher.Hasher
	Metrics    *prometheus_metrics.Metrics
}

//nolint:revive
func NewAuth(s storage.Auth, l logging.Logger, t jwtmanager.TokenMaker, h hasher.Hasher, m *prometheus_metrics.Metrics) Auth {
	return auth{
		Storage:    s,
		Logger:     l,
		TokenMaker: t,
		Hasher:     h,
		Metrics:    m,
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
		s.Logger.Infof("create new user")
		s.Metrics.RegisteredUsers.Inc()
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
		s.Logger.Warnf(
			domain.ErrUnauthorized,
			"wrong password: %v",
			hashedUser.Name,
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

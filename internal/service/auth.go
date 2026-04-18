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
	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	Auth(ctx context.Context, data dto.UserData) (dto.AuthResponseBody, error)
}

type auth struct {
	Storage storage.Auth
	Logger  logging.Logger
}

func NewAuth(s storage.Auth, l logging.Logger) Auth {
	return auth{
		Storage: s,
		Logger:  l,
	}
}

func (s auth) Auth(ctx context.Context, data dto.UserData) (dto.AuthResponseBody, error) {
	hashedUser, err := domain.NewHashed(data.Name, data.Password, s.Logger)
	if err != nil {
		return dto.AuthResponseBody{}, err
	}

	DBHashedPassword, err := s.Storage.GetHashedUserPassword(ctx, hashedUser.Name)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		DBHashedPassword, err = s.Storage.CreateUser(ctx, hashedUser)
		if err != nil {
			return dto.AuthResponseBody{}, fmt.Errorf(
				"failed to create new user: %v",
				err,
			)
		}
		s.Logger.Info("create new user")
	case err != nil:
		return dto.AuthResponseBody{}, fmt.Errorf(
			"failed to get user password from Storage: %v",
			err,
		)

	}

	if err = bcrypt.CompareHashAndPassword(
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

		return dto.AuthResponseBody{}, domain.ErrUnauthorized
	}

	userClaims := domain.DefaultUser{UserName: hashedUser.Name}
	token, err := tools.CreateToken(userClaims, s.Logger)
	if err != nil {
		return dto.AuthResponseBody{}, err
	}

	return dto.AuthResponseBody{Token: token}, nil
}

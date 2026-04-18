package service

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/storage"
	"avito-shop/internal/tools"
	"context"
	"fmt"

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

	DBHashedPassword, isNew, err := s.Storage.GetHashedUserPassword(ctx, hashedUser.Name)
	DBHashedPassword = hashedUser.Password //TODO убрать после добавления репозитория с паролями
	switch {
	case isNew:
		//TODO создание пользователя в БД
		s.Logger.Info(
			fmt.Sprintf(
				"new user created: %v",
				hashedUser.Name,
			),
		)
	case err != nil:
		// TODO Логика с типом ошибок
		switch {
		case false:
		default:
			s.Logger.Error(
				fmt.Sprintf(
					"failed to get user password from Storage, username: %v",
					hashedUser.Name,
				),
				err,
			)
		}

		return dto.AuthResponseBody{}, err
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

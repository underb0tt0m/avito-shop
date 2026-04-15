package service

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/storage"
	"avito-shop/internal/tools"
	"context"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	Auth(ctx context.Context, data dto.UserData) (dto.AuthResponseBody, error)
}

type auth struct {
	Storage storage.Auth
	Logger  *zap.Logger
}

func NewAuth(s storage.Auth, l *zap.Logger) Auth {
	return auth{
		Storage: s,
		Logger:  l,
	}
}

func (s auth) Auth(ctx context.Context, data dto.UserData) (dto.AuthResponseBody, error) {
	hashedUser, err := domain.NewHashed(data.Name, data.Password)
	if err != nil {
		s.Logger.Error(
			"failed to hash password",
			zap.String("username", data.Name),
			zap.Error(err),
		)
		return dto.AuthResponseBody{}, err
	}

	DBHashedPassword, isNew, err := s.Storage.GetHashedUserPassword(ctx, hashedUser.Name)
	DBHashedPassword = hashedUser.Password //TODO убрать после добавления репозитория с паролями
	switch {
	case isNew:
		//TODO создание пользователя в БД
		s.Logger.Info(
			"new user created",
			zap.String("username", hashedUser.Name),
		)
	case err != nil:
		s.Logger.Error(
			"failed to get user password from Storage",
			zap.String("username", hashedUser.Name),
			zap.Error(err),
		)
		return dto.AuthResponseBody{}, err
	}

	if err = bcrypt.CompareHashAndPassword(
		DBHashedPassword,
		[]byte(data.Password),
	); err != nil {
		s.Logger.Warn(
			"wrong password",
			zap.String("username", hashedUser.Name),
			zap.ByteString("your", hashedUser.Password),
			zap.ByteString("true", DBHashedPassword),
			zap.Error(err),
		)

		return dto.AuthResponseBody{}, err
	}

	userClaims := domain.DefaultUser{UserName: hashedUser.Name}
	token, err := tools.CreateToken(userClaims)
	if err != nil {
		s.Logger.Error(
			"failed to generate token",
			zap.String("username", hashedUser.Name),
			zap.Error(err),
		)
		return dto.AuthResponseBody{}, err
	}

	return dto.AuthResponseBody{Token: token}, nil
}

package authService

import (
	"avito-shop/internal/core/domains/domainJwt"

	"avito-shop/internal/features/auth/authRepository"
	"avito-shop/internal/features/auth/authTransport/authDTO"
	"avito-shop/internal/tools"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type ServiceImpl struct {
	Repo   authRepository.Storage
	Logger *zap.Logger
}

func (s ServiceImpl) Auth(data authDTO.UserData) (authDTO.ResponseBody, error) {
	hashedUser, err := domain.NewHashed(data.Name, data.Password)
	if err != nil {
		s.Logger.Error(
			"failed to hash password",
			zap.String("username", data.Name),
			zap.Error(err),
		)
		return authDTO.ResponseBody{}, err
	}

	DBHashedPassword, isNew, err := s.Repo.GetHashedUserPassword(hashedUser.Name, hashedUser.Password)
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
		return authDTO.ResponseBody{}, err
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

		return authDTO.ResponseBody{}, err
	}

	userClaims := domainJwt.DefaultUser{UserName: hashedUser.Name}
	token, err := tools.CreateToken(userClaims)
	if err != nil {
		s.Logger.Error(
			"failed to generate token",
			zap.String("username", hashedUser.Name),
			zap.Error(err),
		)
		return authDTO.ResponseBody{}, err
	}

	return authDTO.ResponseBody{Token: token}, nil
}

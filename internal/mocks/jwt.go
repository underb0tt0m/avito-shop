package mocks

import (
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/tools"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type mockJWTTokenMaker struct {
	CreateTokenFunc       func(data any, logger logging.Logger) (string, error)
	ValidateUserTokenFunc func(tokenString string, logger logging.Logger) error
	ParseUserTokenRawFunc func(tokenString string, logger logging.Logger) ([]byte, error)
}

func NewToken(
	CreateTokenFunc func(data any, logger logging.Logger) (string, error),
	ValidateUserTokenFunc func(tokenString string, logger logging.Logger) error,
	ParseUserTokenRawFunc func(tokenString string, logger logging.Logger) ([]byte, error),
) tools.TokenMaker {
	return mockJWTTokenMaker{
		CreateTokenFunc:       CreateTokenFunc,
		ValidateUserTokenFunc: ValidateUserTokenFunc,
		ParseUserTokenRawFunc: ParseUserTokenRawFunc,
	}
}

func (t mockJWTTokenMaker) CreateToken(data any, logger logging.Logger) (string, error) {
	if t.CreateTokenFunc != nil {
		return t.CreateTokenFunc(data, logger)
	}
	return "test", nil
}

func (t mockJWTTokenMaker) ValidateUserToken(tokenString string, logger logging.Logger) error {
	if t.ValidateUserTokenFunc != nil {
		return t.ValidateUserTokenFunc(tokenString, logger)
	}
	return nil
}

func (t mockJWTTokenMaker) ParseUserTokenRaw(tokenString string, logger logging.Logger) ([]byte, error) {
	if t.ParseUserTokenRawFunc != nil {
		return t.ParseUserTokenRawFunc(tokenString, logger)
	}
	return []byte("test"), nil
}

func createKeyFunc(logger logging.Logger) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error(
				fmt.Sprintf(
					"unexpected signing method: %v",
					token.Header["alg"],
				),
				domain.ErrWrongSigningMethod,
			)
			return nil, domain.ErrWrongSigningMethod
		}
		return config.App.Security.JWTToken.SecretKey, nil
	}
}

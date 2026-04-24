package tools

import (
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenMaker interface {
	CreateToken(data any) (string, error)
	ValidateUserToken(tokenString string) error
	ParseUserTokenRaw(tokenString string) ([]byte, error)
}

type jwtTokenMaker struct {
	logger logging.Logger
}

func NewToken(logger logging.Logger) TokenMaker {
	return jwtTokenMaker{logger}
}

func (t jwtTokenMaker) CreateToken(data any) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.logger.Error(
			"failed to marshal token data",
			err,
		)
		return "", err
	}

	var mapClaims jwt.MapClaims
	if err = json.Unmarshal(
		jsonBytes,
		&mapClaims,
	); err != nil {
		t.logger.Error(
			"failed to unmarshal token data",
			err,
		)
		return "", err
	}

	mapClaims["exp"] = jwt.NewNumericDate(time.Now().Add(config.App.Security.JWTToken.Lifetime))
	mapClaims["iat"] = jwt.NewNumericDate(time.Now())
	mapClaims["iss"] = "app"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	tokenString, err := token.SignedString(config.App.Security.JWTToken.SecretKey)
	if err != nil {
		t.logger.Error(
			"failed to sign token",
			err,
		)
		return "", err
	}
	return tokenString, nil
}

func (t jwtTokenMaker) ValidateUserToken(tokenString string) error {
	_, err := jwt.Parse(tokenString, createKeyFunc(t.logger))
	if err != nil {
		t.logger.Error(
			"invalid token",
			domain.ErrInvalidToken,
		)
		return domain.ErrInvalidToken
	}
	return nil
}

func (t jwtTokenMaker) ParseUserTokenRaw(tokenString string) ([]byte, error) {
	token, err := jwt.Parse(tokenString, createKeyFunc(t.logger))
	if err != nil {
		t.logger.Warn(
			"invalid token",
			domain.ErrInvalidToken,
		)
		return nil, domain.ErrInvalidToken
	}
	if !token.Valid {
		t.logger.Warn(
			"invalid token",
			domain.ErrInvalidToken,
		)
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.logger.Warn(
			"invalid token",
			domain.ErrInvalidToken,
		)
		return nil, domain.ErrInvalidToken
	}
	return json.Marshal(claims)
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

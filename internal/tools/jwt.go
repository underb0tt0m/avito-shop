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

func CreateToken(data any, logger logging.Logger) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logger.Error(
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
		logger.Error(
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
		logger.Error(
			"failed to sign token",
			err,
		)
		return "", err
	}
	return tokenString, nil
}

func ValidateUserToken(tokenString string, logger logging.Logger) error {
	_, err := jwt.Parse(tokenString, createKeyFunc(logger))
	if err != nil {
		logger.Error(
			"invalid token",
			domain.ErrInvalidToken,
		)
		return err
	}
	return nil
}

func ParseUserTokenRaw(tokenString string, logger logging.Logger) ([]byte, error) {
	token, err := jwt.Parse(tokenString, createKeyFunc(logger))
	if err != nil {
		logger.Warn(
			"invalid token",
			domain.ErrInvalidToken,
		)
		return nil, domain.ErrInvalidToken
	}
	if !token.Valid {
		logger.Warn(
			"invalid token",
			domain.ErrInvalidToken,
		)
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logger.Warn(
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

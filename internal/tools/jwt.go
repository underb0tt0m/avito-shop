package tools

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenMaker interface {
	CreateToken(data any) (string, error)
	ValidateUserToken(tokenString string) error
	ParseUserTokenRaw(tokenString string) ([]byte, error)
	GetPrefix() string
}

type jwtTokenMaker struct {
	logger      logging.Logger
	jsonCodec   JSONCodec
	tokenPrefix string
	lifetime    time.Duration
	secret      []byte
}

func NewToken(
	logger logging.Logger,
	jsonCodec JSONCodec,
	tokenPrefix string,
	lifetime time.Duration,
	secret []byte,
) TokenMaker {
	return jwtTokenMaker{
		logger:      logger,
		jsonCodec:   jsonCodec,
		tokenPrefix: tokenPrefix,
		lifetime:    lifetime,
		secret:      secret,
	}
}

func (t jwtTokenMaker) CreateToken(data any) (string, error) {
	jsonBytes, err := t.jsonCodec.Marshal(data)
	if err != nil {
		t.logger.Error(
			"failed to marshal token data",
			err,
		)
		return "", err
	}

	var mapClaims jwt.MapClaims
	if err = t.jsonCodec.Unmarshal(
		jsonBytes,
		&mapClaims,
	); err != nil {
		t.logger.Error(
			"failed to unmarshal token data",
			err,
		)
		return "", err
	}

	mapClaims["exp"] = jwt.NewNumericDate(time.Now().Add(t.lifetime))
	mapClaims["iat"] = jwt.NewNumericDate(time.Now())
	mapClaims["iss"] = "app"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	tokenString, err := token.SignedString(t.secret)
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
	_, err := jwt.Parse(tokenString, createKeyFunc(t.logger, t.secret))
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
	token, err := jwt.Parse(tokenString, createKeyFunc(t.logger, t.secret))
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
	return t.jsonCodec.Marshal(claims)
}

func (t jwtTokenMaker) GetPrefix() string {
	return t.tokenPrefix
}

func createKeyFunc(logger logging.Logger, secret []byte) jwt.Keyfunc {
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
		return secret, nil
	}
}

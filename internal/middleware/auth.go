package middleware

import (
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/tools"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func Auth(r *http.Request, logger logging.Logger) (*domain.DefaultUser, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		logger.Warn(
			"user unauthorized",
			domain.ErrUnauthorized,
		)
		return nil, domain.ErrUnauthorized
	}
	token, ok := strings.CutPrefix(token, config.App.Security.JWTToken.Prefix)
	if !ok {
		logger.Warn(
			"Token without prefix",
			domain.ErrInvalidToken,
		)
		return nil, domain.ErrInvalidToken
	}
	token = strings.TrimSpace(token)
	jsonBytes, err := tools.ParseUserTokenRaw(token, logger)
	if err != nil {
		return nil, err
	}
	var claims domain.DefaultUser
	if err = json.Unmarshal(
		jsonBytes,
		&claims,
	); err != nil {
		logger.Error(
			"failed to unmarshal token",
			err,
		)
		return nil, domain.ErrBadRequest
	}

	if claims.ExpiresAt.Unix() < time.Now().Unix() {
		logger.Warn(
			"token is expired",
			domain.ErrTokenExpired,
		)
		return nil, domain.ErrTokenExpired
	}
	return &claims, nil
}

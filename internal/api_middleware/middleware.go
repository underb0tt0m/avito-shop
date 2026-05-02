package api_middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"avito-shop/internal/domain"
	"avito-shop/internal/jwtmanager"
	"avito-shop/internal/logging"
)

func Stopwatch(logger logging.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start).Milliseconds()
			logger.Infof(
				"method: %v, path: %v, address: %v, duration: %v ms",
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				duration,
			)
		})
	}
}

func Auth(logger logging.Logger, tokenMaker jwtmanager.TokenMaker) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				logger.Warnf(
					domain.ErrUnauthorized,
					"user unauthorized",
				)
				domain.WriteError(w, domain.ErrUnauthorized, logger)
				return
			}
			token, ok := strings.CutPrefix(token, tokenMaker.GetPrefix())
			if !ok {
				logger.Warnf(
					domain.ErrInvalidToken,
					"Token without prefix",
				)
				domain.WriteError(w, domain.ErrInvalidToken, logger)
				return
			}
			token = strings.TrimSpace(token)
			jsonBytes, err := tokenMaker.ParseUserTokenRaw(token)
			if err != nil {
				domain.WriteError(w, err, logger)
				return
			}
			var claims domain.DefaultUser
			if err = json.Unmarshal(
				jsonBytes,
				&claims,
			); err != nil {
				logger.Errorf(
					err,
					"failed to unmarshal token",
				)
				domain.WriteError(w, domain.ErrBadRequest, logger)
				return
			}

			if claims.ExpiresAt.Unix() < time.Now().Unix() {
				logger.Warnf(
					domain.ErrTokenExpired,
					"token is expired",
				)
				domain.WriteError(w, domain.ErrTokenExpired, logger)
				return
			}

			ctx := context.WithValue(r.Context(), domain.UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

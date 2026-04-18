package api_middleware

import (
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/tools"
	"avito-shop/internal/tools/consts"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func Stopwatch(logger logging.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start).Milliseconds()
			logger.Info(
				fmt.Sprintf(
					"method: %v, path: %v, address: %v, duration: %v ms",
					r.Method,
					r.URL.Path,
					r.RemoteAddr,
					duration,
				),
			)
		})
	}
}

func Auth(logger logging.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				logger.Warn(
					"user unauthorized",
					domain.ErrUnauthorized,
				)
				tools.WriteError(w, domain.ErrUnauthorized)
				return
			}
			token, ok := strings.CutPrefix(token, config.App.Security.JWTToken.Prefix)
			if !ok {
				logger.Warn(
					"Token without prefix",
					domain.ErrInvalidToken,
				)
				tools.WriteError(w, domain.ErrInvalidToken)
				return
			}
			token = strings.TrimSpace(token)
			jsonBytes, err := tools.ParseUserTokenRaw(token, logger)
			if err != nil {
				tools.WriteError(w, err)
				return
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
				tools.WriteError(w, domain.ErrBadRequest)
				return
			}

			if claims.ExpiresAt.Unix() < time.Now().Unix() {
				logger.Warn(
					"token is expired",
					domain.ErrTokenExpired,
				)
				tools.WriteError(w, domain.ErrTokenExpired)
				return
			}

			ctx := context.WithValue(r.Context(), consts.UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package authTransport

import (
	"avito-shop/internal/features/auth/authService"
	"avito-shop/internal/features/auth/authTransport/authDTO"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Register(s authService.Service, r chi.Router, logger *zap.Logger) {
	r.Post("/auth", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Debug(
			"request received",
			zap.String("method", r.Method),
			zap.String("pattern", r.Pattern),
			zap.String("remoteAddr", r.RemoteAddr),
		)

		requestBody, err := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()
		if err != nil {
			logger.Error(
				"failed to read auth request body",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user := authDTO.UserData{}
		if err = json.Unmarshal(
			requestBody,
			&user,
		); err != nil {
			logger.Error(
				"failed to unmarshal auth request body",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := s.Auth(user)
		if err != nil {
			logger.Warn(
				"authentication denied",
				zap.String("username", user.Name),
			)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		responseBody, err := json.Marshal(token)
		if err != nil {
			logger.Error(
				"failed to marshal auth response body",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err = w.Write(responseBody); err != nil {
			logger.Error(
				"failed to write auth response",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Debug(
			"request has been processed",
			zap.Int("status", http.StatusOK),
			zap.Duration("processing time", time.Since(start)),
		)
	})
}

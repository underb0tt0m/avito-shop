package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/config"
	"avito-shop/internal/service"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Auth(s service.Auth, r chi.Router, logger *zap.Logger) {
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

		user := dto.UserData{}
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

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		logger.Debug(
			"calling AuthService Auth method",
			zap.String("username", user.Name),
		)
		token, err := s.Auth(ctx, user)
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

package transport

import (
	"avito-shop/internal/features/api/service"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const queryTimeout = time.Second

func Register(s service.Service, logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/api/info", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Debug(
			"request received",
			zap.String("method", r.Method),
			zap.String("pattern", r.Pattern),
			zap.String("remoteAddr", r.RemoteAddr),
		)

		token := r.Header.Get("Authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			logger.Warn(
				"missing authorization token",
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
			)
			return
		}
		logger.Debug("JWT token received")
		//username := ParseJWT(token) TODO после добавления авторизации
		username := "Artem"

		ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
		defer cancel()
		logger.Debug(
			"calling service GetUserInfo method",
			zap.String("username", username),
		)
		dtoUser, err := s.GetUserInfo(ctx, username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			//TODO реализовать логику с вытягиванием статуса и тела ошибки
			logger.Error(
				"failed to get user info",
				zap.Error(err),
				zap.String("username", username),
			)
			return
		}

		response, err := json.MarshalIndent(dtoUser, "", "	")
		if err != nil {
			logger.Error(
				"failed to marshal user info response",
				zap.Error(err),
				zap.String("username", username),
			)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if _, err = w.Write(response); err != nil {
			logger.Error(
				"failed to write response",
				zap.Error(err),
				zap.String("username", username),
			)
			w.WriteHeader(http.StatusInternalServerError)
		}
		logger.Debug(
			"request has been processed",
			zap.Int("status", http.StatusOK),
			zap.Duration("processing time", time.Since(start)),
		)
	})

	return r
}

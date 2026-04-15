package handler

import (
	"avito-shop/internal/config"
	"avito-shop/internal/middleware"
	"avito-shop/internal/service"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Main(s service.API, r chi.Router, logger *zap.Logger) {
	r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Debug(
			"request received",
			zap.String("method", r.Method),
			zap.String("pattern", r.Pattern),
			zap.String("remoteAddr", r.RemoteAddr),
		)

		claims, err := middleware.Auth(w, r)
		if err != nil {
			switch err.Error() {
			case "unauthorized":
				logger.Warn("user is unauthorized")
				w.WriteHeader(http.StatusUnauthorized)
				return
			case "token expired":
				logger.Info(
					"user's token has expired",
					zap.String("username", claims.UserName),
				)
				w.WriteHeader(http.StatusUnauthorized)
				return
			case "token is malformed: token contains an invalid number of segment":
				logger.Debug("bad request")
				w.WriteHeader(http.StatusBadRequest)
				return
			default:
				logger.Info(
					"failed to validate user's token",
					zap.Error(err),
				)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		logger.Debug("JWT token is valid")
		username := claims.UserName

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		logger.Debug(
			"calling mainRoutService GetUserInfo method",
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
				"failed to write info response",
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
}

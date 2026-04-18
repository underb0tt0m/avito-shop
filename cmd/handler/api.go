package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/middleware"
	"avito-shop/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func Main(s service.API, r chi.Router, logger logging.Logger) {
	r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Debug(
			fmt.Sprintf(
				"request received, method: %v, pattern: %v, remoteAddr: %v",
				r.Method,
				r.Pattern,
				r.RemoteAddr,
			),
		)

		claims, err := middleware.Auth(w, r, logger)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrUnauthorized):
				w.WriteHeader(domain.ErrUnauthorized.Code)
				return
			case errors.Is(err, domain.ErrTokenExpired):
				w.WriteHeader(domain.ErrTokenExpired.Code)
				return
			case errors.Is(err, domain.ErrBadRequest):
				w.WriteHeader(domain.ErrBadRequest.Code)
				return
			default:
				w.WriteHeader(domain.ErrInternalServerError.Code)
				return
			}
		}

		logger.Debug("JWT token is valid")
		username := claims.UserName

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		logger.Debug("calling mainRoutService GetUserInfo method")
		dtoUser, err := s.GetUserInfo(ctx, username)
		if err != nil {
			//TODO реализовать логику с вытягиванием статуса и тела ошибки
			switch {
			case false:
			default:
				w.WriteHeader(domain.ErrInternalServerError.Code)
			}
			return
		}

		response, err := json.MarshalIndent(dtoUser, "", "	")
		if err != nil {
			logger.Error(
				"failed to marshal user info response",
				err,
			)
			w.WriteHeader(domain.ErrInternalServerError.Code)
		}
		if _, err = w.Write(response); err != nil {
			logger.Error(
				"failed to write info response",
				err,
			)
			w.WriteHeader(domain.ErrInternalServerError.Code)
		}
		logger.Debug(
			fmt.Sprintf(
				"request has been processed, status: %v, processing time: %v",
				http.StatusOK,
				time.Since(start),
			),
		)
	})

	r.Post("/sendCoin", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Debug(
			fmt.Sprintf(
				"request received, method: %v, pattern: %v, remoteAddr: %v",
				r.Method,
				r.Pattern,
				r.RemoteAddr,
			),
		)

		requestBody, err := io.ReadAll(r.Body)
		defer func() { _ = r.Body.Close() }()
		if err != nil {
			logger.Error(
				"failed to read request body",
				err,
			)
			w.WriteHeader(domain.ErrInternalServerError.Code)
			return
		}

		transaction := dto.SendCoinRequest{}
		if err = json.Unmarshal(
			requestBody,
			&transaction,
		); err != nil {
			logger.Error(
				"failed to unmarshal request body",
				err,
			)
			w.WriteHeader(domain.ErrInternalServerError.Code)
			return
		}

		sender, err := middleware.Auth(w, r, logger)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidToken):
				w.WriteHeader(domain.ErrInvalidToken.Code)
			case errors.Is(err, domain.ErrUnauthorized):
				w.WriteHeader(domain.ErrUnauthorized.Code)
			case errors.Is(err, domain.ErrBadRequest):
				w.WriteHeader(domain.ErrBadRequest.Code)
			case errors.Is(err, domain.ErrTokenExpired):
				w.WriteHeader(domain.ErrTokenExpired.Code)
			default:
				w.WriteHeader(domain.ErrInternalServerError.Code)
			}
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		if err = s.SendCoins(
			ctx,
			sender.UserName,
			transaction,
		); err != nil {
			switch {
			case errors.Is(err, domain.ErrInsufficientFunds):
				w.WriteHeader(domain.ErrInsufficientFunds.Code)
			case errors.Is(err, domain.ErrNotFound):
				w.WriteHeader(domain.ErrNotFound.Code)
			case errors.Is(err, domain.ErrBadRequest):
				w.WriteHeader(domain.ErrBadRequest.Code)
			default:
				w.WriteHeader(domain.ErrInternalServerError.Code)
			}
			return
		}

		logger.Debug(
			fmt.Sprintf(
				"request has been processed, status: %v, processing time: %v",
				http.StatusOK,
				time.Since(start),
			),
		)
	})

	r.Post("/buy/{item}", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Debug(
			fmt.Sprintf(
				"request received, method: %v, pattern: %v, remoteAddr: %v",
				r.Method,
				r.Pattern,
				r.RemoteAddr,
			),
		)

		user, err := middleware.Auth(w, r, logger)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidToken):
				w.WriteHeader(domain.ErrInvalidToken.Code)
			case errors.Is(err, domain.ErrUnauthorized):
				w.WriteHeader(domain.ErrUnauthorized.Code)
			case errors.Is(err, domain.ErrBadRequest):
				w.WriteHeader(domain.ErrBadRequest.Code)
			case errors.Is(err, domain.ErrTokenExpired):
				w.WriteHeader(domain.ErrTokenExpired.Code)
			default:
				w.WriteHeader(domain.ErrInternalServerError.Code)
			}
			return
		}

		strItemID := chi.URLParam(r, "item")
		if strItemID == "" {
			logger.Warn(
				"attempt to buy item with empty param {item}",
				domain.ErrBadRequest,
			)
			w.WriteHeader(domain.ErrBadRequest.Code)
			return
		}

		itemID, err := strconv.Atoi(strItemID)
		if err != nil {
			logger.Warn(
				"attempt to buy item with invalid id",
				domain.ErrBadRequest,
			)
			w.WriteHeader(domain.ErrBadRequest.Code)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		if err = s.BuyItem(ctx, itemID, user.UserName); err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				w.WriteHeader(domain.ErrNotFound.Code)
			case errors.Is(err, domain.ErrInsufficientFunds):
				w.WriteHeader(domain.ErrInsufficientFunds.Code)
			case errors.Is(err, domain.ErrInternalServerError):
				w.WriteHeader(domain.ErrInternalServerError.Code)
			default:
				w.WriteHeader(domain.ErrInternalServerError.Code)
			}
			return
		}

		logger.Debug(
			fmt.Sprintf(
				"request has been processed, status: %v, processing time: %v",
				http.StatusOK,
				time.Since(start),
			),
		)
	})
}

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

		claims, err := middleware.Auth(r, logger)
		if err != nil {
			WriteError(w, err)
			return
		}

		logger.Debug("JWT token is valid")
		username := claims.UserName

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		logger.Debug("calling mainRoutService GetUserInfo method")
		dtoUser, err := s.GetUserInfo(ctx, username)
		if err != nil {
			WriteError(w, err)
			return
		}

		response, err := json.MarshalIndent(dtoUser, "", "	")
		if err != nil {
			logger.Error(
				"failed to marshal user info response",
				err,
			)
			WriteError(w, err)
			return
		}
		if _, err = w.Write(response); err != nil {
			logger.Error(
				"failed to write info response",
				err,
			)
			WriteError(w, err)
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
			WriteError(w, err)
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
			WriteError(w, err)
			return
		}

		sender, err := middleware.Auth(r, logger)
		if err != nil {
			WriteError(w, err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		if err = s.SendCoins(
			ctx,
			sender.UserName,
			transaction,
		); err != nil {
			WriteError(w, err)
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

	r.Post("/buy/{itemID}", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Debug(
			fmt.Sprintf(
				"request received, method: %v, pattern: %v, remoteAddr: %v",
				r.Method,
				r.Pattern,
				r.RemoteAddr,
			),
		)

		user, err := middleware.Auth(r, logger)
		if err != nil {
			WriteError(w, err)
			return
		}

		strItemID := chi.URLParam(r, "itemID")
		if strItemID == "" {
			logger.Warn(
				"attempt to buy item with empty param {item}",
				domain.ErrBadRequest,
			)
			WriteError(w, domain.ErrBadRequest)
			return
		}

		itemID, err := strconv.Atoi(strItemID)
		if err != nil {
			logger.Warn(
				"attempt to buy item with invalid id",
				domain.ErrBadRequest,
			)
			WriteError(w, domain.ErrBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), config.App.Storage.QueryTimeout)
		defer cancel()
		if err = s.BuyItem(ctx, itemID, user.UserName); err != nil {
			WriteError(w, err)
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

func WriteError(w http.ResponseWriter, err error) {
	if apiErr, ok := errors.AsType[domain.APIErr](err); ok {
		w.WriteHeader(apiErr.Code)
		_, _ = w.Write([]byte(apiErr.Message))
		return
	}
	w.WriteHeader(domain.ErrInternalServerError.Code)
	_, _ = w.Write([]byte(domain.ErrInternalServerError.Message))
	return
}

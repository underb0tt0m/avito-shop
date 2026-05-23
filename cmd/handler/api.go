package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"avito-shop/cmd/dto"
	"avito-shop/cmd/handler/http_error"
	"avito-shop/internal/api_middleware"
	"avito-shop/internal/domain"
	"avito-shop/internal/jsoncodec"
	"avito-shop/internal/jwtmanager"
	"avito-shop/internal/logging"
	"avito-shop/internal/prometheus_metrics"
	"avito-shop/internal/service"

	"github.com/go-chi/chi/v5"
)

type Main struct {
	service      service.API
	logger       logging.Logger
	jsonCodec    jsoncodec.JSONCodec
	queryTimeout time.Duration
	tokenMaker   jwtmanager.TokenMaker
	metrics      *prometheus_metrics.Metrics
}

//nolint:revive
func NewMain(
	service service.API,
	logger logging.Logger,
	jsonCodec jsoncodec.JSONCodec,
	queryTimeout time.Duration,
	tokenMaker jwtmanager.TokenMaker,
	metrics *prometheus_metrics.Metrics,
) Main {
	return Main{
		service:      service,
		logger:       logger,
		jsonCodec:    jsonCodec,
		queryTimeout: queryTimeout,
		tokenMaker:   tokenMaker,
		metrics:      metrics,
	}
}

func (h Main) RegisterRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(api_middleware.Auth(h.logger, h.tokenMaker, h.metrics))

		h.GetInfo(r)
		h.SendCoin(r)
		h.BuyItem(r)
	})
}

func (h Main) GetInfo(r chi.Router) {
	r.Get("/info", h.handleInfo)
}

func (h Main) handleInfo(w http.ResponseWriter, r *http.Request) {
	user, err := domain.GetUserFromContext(r)
	if err != nil {
		http_error.Write(w, err, h.logger, h.metrics)
		return
	}

	h.logger.Debugf("JWT token is valid")
	username := user.UserName

	ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
	defer cancel()
	h.logger.Debugf("calling mainRoutService GetUserInfo method")
	dtoUser, err := h.service.GetUserInfo(ctx, username)
	if err != nil {
		http_error.Write(w, err, h.logger, h.metrics)
		return
	}

	response, err := h.jsonCodec.MarshalIndent(dtoUser, "", "	")
	if err != nil {
		h.logger.Errorf(
			"failed to marshal user info response: %v",
			err,
		)
		http_error.Write(w, err, h.logger, h.metrics)
		return
	}
	if _, err = w.Write(response); err != nil {
		h.logger.Errorf(
			"failed to write info response: %v",
			err,
		)
		http_error.Write(w, err, h.logger, h.metrics)
		return
	}
}

func (h Main) SendCoin(r chi.Router) {
	r.Post("/sendCoin", h.handleSendCoin)
}

func (h Main) handleSendCoin(w http.ResponseWriter, r *http.Request) {
	requestBody, err := io.ReadAll(r.Body)
	defer func() {
		if err = r.Body.Close(); err != nil {
			h.logger.Errorf(
				"failed to close request body: %v",
				err,
			)
		}
	}()
	if err != nil {
		h.logger.Errorf(
			"failed to read request body: %v",
			err,
		)
		http_error.Write(w, err, h.logger, h.metrics)
		return
	}

	transaction := dto.SendCoinRequest{}
	if err = h.jsonCodec.Unmarshal(
		requestBody,
		&transaction,
	); err != nil {
		h.logger.Errorf(
			"failed to unmarshal request body: %v",
			err,
		)
		http_error.Write(w, domain.ErrUnprocessableEntity, h.logger, h.metrics)
		return
	}

	user, err := domain.GetUserFromContext(r)
	if err != nil {
		http_error.Write(w, err, h.logger, h.metrics)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
	defer cancel()
	if err = h.service.SendCoins(
		ctx,
		user.UserName,
		transaction,
	); err != nil {
		http_error.Write(w, err, h.logger, h.metrics)
		h.metrics.Transactions.WithLabelValues("app:8080", prometheus_metrics.StatusTransactionFailed).Inc()
		return
	}
	h.metrics.Transactions.WithLabelValues("app:8080", prometheus_metrics.StatusTransactionSuccess).Inc()
}

func (h Main) BuyItem(r chi.Router) {
	r.Post("/buy/{itemID}", h.handleBuyItem)
}

func (h Main) handleBuyItem(w http.ResponseWriter, r *http.Request) {
	user, err := domain.GetUserFromContext(r)
	if err != nil {
		http_error.Write(w, err, h.logger, h.metrics)
		return
	}

	strItemID := chi.URLParam(r, "itemID")
	if strItemID == "" {
		h.logger.Warnf(
			"attempt to buy item with empty param {item}: %v",
			domain.ErrBadRequest,
		)
		http_error.Write(w, domain.ErrBadRequest, h.logger, h.metrics)
		return
	}

	itemID, err := strconv.Atoi(strItemID)
	if err != nil {
		h.logger.Warnf(
			"attempt to buy item with invalid id: %v",
			domain.ErrBadRequest,
		)
		http_error.Write(w, domain.ErrBadRequest, h.logger, h.metrics)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
	defer cancel()
	if err = h.service.BuyItem(ctx, itemID, user.UserName); err != nil {
		http_error.Write(w, err, h.logger, h.metrics)
		return
	}
	h.metrics.Purchases.WithLabelValues("app:8080").Inc()
}

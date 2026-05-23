package http_error

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/prometheus_metrics"
)

func Write(w http.ResponseWriter, err error, logger logging.Logger, m *prometheus_metrics.Metrics) {
	if apiErr, ok := errors.AsType[domain.APIErr](err); ok {
		w.WriteHeader(apiErr.Code)

		m.Errors.WithLabelValues("app:8080", strconv.Itoa(apiErr.Code)).Inc()

		response, marshalErr := json.Marshal(dto.ErrorResponse{Errors: apiErr.Message})
		if marshalErr != nil {
			logger.Errorf("failed to marshal request body: %v", err)
			return
		}
		if _, err = w.Write(response); err != nil {
			logger.Errorf("failed to write request body: %v", err)
			return
		}
		return
	}
	m.Errors.WithLabelValues("app:8080", strconv.Itoa(domain.ErrInternalServerError.Code)).Inc()
	w.WriteHeader(domain.ErrInternalServerError.Code)
	response, err := json.Marshal(dto.ErrorResponse{Errors: domain.ErrInternalServerError.Message})
	if err != nil {
		logger.Errorf("failed to marshal response body: %v", err)
		return
	}
	_, err = w.Write(response)
	if err != nil {
		logger.Errorf("failed to write response body: %v", err)
		return
	}
}

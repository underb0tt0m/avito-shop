package domain

import (
	"encoding/json"
	"errors"
	"net/http"

	"avito-shop/cmd/dto"
	"avito-shop/internal/logging"
)

type APIErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e APIErr) Error() string {
	return e.Message
}

var (
	ErrNotFound = APIErr{
		Code:    http.StatusNotFound,
		Message: "user or item not found",
	}
	ErrBadRequest = APIErr{
		Code:    http.StatusBadRequest,
		Message: "invalid request format or parameters",
	}
	ErrInternalServerError = APIErr{
		Code:    http.StatusInternalServerError,
		Message: "internal server error, please try again later",
	}
	ErrUnauthorized = APIErr{
		Code:    http.StatusUnauthorized,
		Message: "authorization required",
	}
	ErrInvalidToken = APIErr{
		Code:    http.StatusUnauthorized,
		Message: "invalid or malformed token",
	}
	ErrWrongSigningMethod = APIErr{
		Code:    http.StatusUnauthorized,
		Message: "unsupported token signing method",
	}
	ErrTokenExpired = APIErr{
		Code:    http.StatusUnauthorized,
		Message: "token has expired, please login again",
	}
	ErrInsufficientFunds = APIErr{
		Code:    http.StatusPaymentRequired,
		Message: "insufficient coins balance",
	}
	ErrUnprocessableEntity = APIErr{
		Code:    http.StatusUnprocessableEntity,
		Message: "invalid request body",
	}
)

func WriteError(w http.ResponseWriter, err error, logger logging.Logger) {
	if apiErr, ok := errors.AsType[APIErr](err); ok {
		response, marshalErr := json.Marshal(dto.ErrorResponse{Errors: apiErr.Message})
		if marshalErr != nil {
			logger.Errorf(err, "failed to marshal request body: %v", err)
		}
		w.WriteHeader(apiErr.Code)
		if _, err = w.Write(response); err != nil {
			logger.Errorf(err, "failed to write request body: %v", err)
		}

	}
	response, err := json.Marshal(dto.ErrorResponse{Errors: ErrInternalServerError.Message})
	if err != nil {
		logger.Errorf(err, "failed to marshal response body")
	}
	w.WriteHeader(ErrInternalServerError.Code)
	_, err = w.Write(response)
	if err != nil {
		logger.Errorf(err, "failed to write response body")
	}
}

package domain

import (
	"net/http"
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

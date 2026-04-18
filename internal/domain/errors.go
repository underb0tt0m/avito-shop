package domain

type APIErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e APIErr) Error() string {
	return e.Message
}

var (
	ErrNotFound = APIErr{
		Code:    404,
		Message: "user or item not found",
	}
	ErrBadRequest = APIErr{
		Code:    400,
		Message: "invalid request format or parameters",
	}
	ErrInternalServerError = APIErr{
		Code:    500,
		Message: "internal server error, please try again later",
	}
	ErrUnauthorized = APIErr{
		Code:    401,
		Message: "authorization required",
	}
	ErrInvalidToken = APIErr{
		Code:    401,
		Message: "invalid or malformed token",
	}
	ErrWrongSigningMethod = APIErr{
		Code:    401,
		Message: "unsupported token signing method",
	}
	ErrTokenExpired = APIErr{
		Code:    401,
		Message: "token has expired, please login again",
	}
	ErrInsufficientFunds = APIErr{
		Code:    402,
		Message: "insufficient coins balance",
	}
)

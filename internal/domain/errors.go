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
		Message: "not found",
	}
	ErrBadRequest = APIErr{
		Code:    400,
		Message: "bad request",
	}
	ErrInternalServerError = APIErr{
		Code:    500,
		Message: "internal server error",
	}
)

// if errors.As(err, &APIErr{})

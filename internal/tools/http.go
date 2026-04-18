package tools

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/tools/consts"
	"errors"
	"net/http"
)

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

func GetUserFromContext(r *http.Request) (domain.DefaultUser, error) {
	claims := r.Context().Value(consts.UserContextKey)
	if claims == nil {
		return domain.DefaultUser{}, domain.ErrUnauthorized
	}

	user, ok := claims.(domain.DefaultUser)
	if !ok {
		return domain.DefaultUser{}, domain.ErrInternalServerError
	}

	return user, nil
}

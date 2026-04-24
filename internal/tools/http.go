package tools

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/tools/consts"
	"encoding/json"
	"errors"
	"net/http"
)

func WriteError(w http.ResponseWriter, err error) {
	if apiErr, ok := errors.AsType[domain.APIErr](err); ok {
		response, _ := json.Marshal(dto.ErrorResponse{Errors: apiErr.Message})
		w.WriteHeader(apiErr.Code)
		_, _ = w.Write(response)
		return
	}
	response, _ := json.Marshal(dto.ErrorResponse{Errors: domain.ErrInternalServerError.Message})
	w.WriteHeader(domain.ErrInternalServerError.Code)
	_, _ = w.Write(response)
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

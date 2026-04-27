package domain

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

type DefaultUser struct {
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

func GetUserFromContext(r *http.Request) (DefaultUser, error) {
	claims := r.Context().Value(UserContextKey)
	if claims == nil {
		return DefaultUser{}, ErrUnauthorized
	}

	user, ok := claims.(DefaultUser)
	if !ok {
		return DefaultUser{}, ErrInternalServerError
	}

	return user, nil
}

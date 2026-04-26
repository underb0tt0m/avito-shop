package domain

import "github.com/golang-jwt/jwt/v5"

type contextKey string

const UserContextKey contextKey = "user"

type DefaultUser struct {
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

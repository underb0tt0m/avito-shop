package domainJwt

import "github.com/golang-jwt/jwt/v5"

type DefaultUser struct {
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("JWT_SECRET"))

func CreateToken(data any) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	var mapClaims jwt.MapClaims
	if err = json.Unmarshal(
		jsonBytes,
		&mapClaims,
	); err != nil {
		return "", err
	}

	mapClaims["ExpiresAt"] = jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
	mapClaims["IssuedAt"] = jwt.NewNumericDate(time.Now())
	mapClaims["Issuer"] = "app"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateUserToken(tokenString string) error {
	_, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return err
	}
	return nil
}

func ParseUserTokenRaw(tokenString string) ([]byte, error) {
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to cast claims")
	}
	return json.Marshal(claims)
}

func keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return secretKey, nil
}

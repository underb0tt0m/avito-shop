package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"avito-shop/internal/core/domains/domainJwt"

	"avito-shop/internal/tools"
)

func Auth(w http.ResponseWriter, r *http.Request) (*domainJwt.DefaultUser, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return nil, fmt.Errorf("unauthorized")
	}
	jsonBytes, err := tools.ParseUserTokenRaw(token)
	if err != nil {
		if err.Error() == "invalid token" {
			return nil, fmt.Errorf("unauthorized")
		}
		return nil, err
	}
	var claims domainJwt.DefaultUser
	if err = json.Unmarshal(
		jsonBytes,
		&claims,
	); err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if claims.ExpiresAt.Unix() < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}
	return &claims, nil
}

// metrics, logs, recover

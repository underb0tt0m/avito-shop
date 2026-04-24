package mocks

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/tools"
	"avito-shop/internal/tools/consts"
	"context"
	"net/http"
)

func Auth(logger logging.Logger, tokenMaker tools.TokenMaker) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(
				r.Context(),
				consts.UserContextKey,
				domain.DefaultUser{
					UserName: "test",
				},
			)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

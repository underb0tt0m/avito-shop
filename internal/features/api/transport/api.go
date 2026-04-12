package transport

import (
	"avito-shop/internal/features/api/service"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Register(s service.Service) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/api/info", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			// TODO logger
			return
		}
		//username := ParseJWT(token) TODO после добавления авторизации
		username := "timur"
		dtoUser, err := s.GetUserInfo(username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			//TODO реализовать логику с вытягиванием статуса и тела ошибки
			// TODO logger
			return
		}

		response, err := json.MarshalIndent(dtoUser, "", "	")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if _, err = w.Write(response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	return r
}

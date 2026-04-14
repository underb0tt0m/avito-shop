package authService

import "avito-shop/internal/features/auth/authTransport/authDTO"

type Service interface {
	Auth(data authDTO.UserData) (authDTO.ResponseBody, error)
}

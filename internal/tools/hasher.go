package tools

import (
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Hash(data string, logger logging.Logger) ([]byte, error)
	CompareHashAndPassword(hashedPassword []byte, password []byte) error
}

type bcryptHasher struct{}

func NewHasher() Hasher {
	return bcryptHasher{}
}

func (h bcryptHasher) Hash(data string, logger logging.Logger) ([]byte, error) {
	hashedData, err := bcrypt.GenerateFromPassword([]byte(data), config.App.Security.Hash.Cost)
	if err != nil {
		logger.Error(
			fmt.Sprintf(
				"failed to hash data",
			),
			err,
		)
		return []byte{}, domain.ErrInternalServerError
	}

	return hashedData, nil
}

func (h bcryptHasher) CompareHashAndPassword(hashedPassword []byte, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

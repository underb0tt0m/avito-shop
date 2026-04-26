package tools

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Hash(data string, logger logging.Logger) ([]byte, error)
	CompareHashAndPassword(hashedPassword []byte, password []byte) error
}

type bcryptHasher struct {
	hashCost int
}

func NewHasher(hashCost int) Hasher {
	return bcryptHasher{hashCost}
}

func (h bcryptHasher) Hash(data string, logger logging.Logger) ([]byte, error) {
	hashedData, err := bcrypt.GenerateFromPassword([]byte(data), h.hashCost)
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

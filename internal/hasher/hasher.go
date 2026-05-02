package hasher

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"

	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=hasher.go -destination=../mocks/hasher.go -package=mocks -mock_names=Hasher=Hasher
type Hasher interface {
	Hash(data string, logger logging.Logger) ([]byte, error)
	CompareHashAndPassword(hashedPassword []byte, password []byte) error
}

type bcryptHasher struct {
	hashCost int
}

func NewHasher(hashCost int) Hasher {
	return bcryptHasher{hashCost: hashCost}
}

func (h bcryptHasher) Hash(data string, logger logging.Logger) ([]byte, error) {
	hashedData, err := bcrypt.GenerateFromPassword([]byte(data), h.hashCost)
	if err != nil {
		logger.Errorf(
			err,
			"failed to hash data",
		)
		return []byte{}, domain.ErrInternalServerError
	}

	return hashedData, nil
}

func (bcryptHasher) CompareHashAndPassword(hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

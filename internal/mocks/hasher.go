package mocks

import (
	"avito-shop/internal/logging"
	"avito-shop/internal/tools"
)

type mockBcryptHasher struct {
	HashFunc                   func(data string, logger logging.Logger) ([]byte, error)
	CompareHashAndPasswordFunc func(hashedPassword []byte, password []byte) error
}

func NewHasher(
	HashFunc func(data string, logger logging.Logger) ([]byte, error),
	CompareHashAndPasswordFunc func(hashedPassword []byte, password []byte) error,
) tools.Hasher {
	return mockBcryptHasher{
		HashFunc:                   HashFunc,
		CompareHashAndPasswordFunc: CompareHashAndPasswordFunc,
	}
}

func (h mockBcryptHasher) Hash(data string, logger logging.Logger) ([]byte, error) {
	if h.HashFunc != nil {
		return h.HashFunc(data, logger)
	}
	return []byte("test"), nil
}

func (h mockBcryptHasher) CompareHashAndPassword(hashedPassword []byte, password []byte) error {
	if h.CompareHashAndPasswordFunc != nil {
		return h.CompareHashAndPasswordFunc(hashedPassword, password)
	}
	return nil
}

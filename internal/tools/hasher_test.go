package tools

import (
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type testLogger struct{}

func (l testLogger) Debug(string)        {}
func (l testLogger) Info(string)         {}
func (l testLogger) Warn(string, error)  {}
func (l testLogger) Error(string, error) {}
func (l testLogger) Fatal(string, error) {}
func (l testLogger) Sync() error         { return nil }

func TestCompareHashAndPassword(t *testing.T) {
	tests := []struct {
		name            string
		data            string
		wantErr         bool
		wantSpecificErr error
	}{
		{
			"successfull_hash_data",
			"test",
			false,
			nil,
		},
	}

	hasher := NewHasher()
	logger := testLogger{}
	if err := config.Init("../../cmd/config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	for _, test := range tests {
		_, err := hasher.Hash(test.data, logger)
		if err != nil && !test.wantErr {
			t.Fatalf("Test %v, Hash() unexpected error: %v", test.name, err)
		}
		if err == nil && test.wantErr {
			t.Fatalf("Test %v, Hash() unhandled error: %v", test.name, err) // хз, как это может быть
		}
	}
}

func TestHash(t *testing.T) {
	tests := []struct {
		name            string
		hashedPassword  []byte
		password        []byte
		wantErr         bool
		wantSpecificErr error
	}{
		{
			"successfull_compare",
			[]byte{},
			[]byte("correct_password"),
			false,
			domain.ErrInternalServerError,
		},

		{
			"error_wrong_password",
			[]byte{},
			[]byte("wrong_password"),
			true,
			nil,
		},
	}

	hasher := NewHasher()
	if err := config.Init("../../cmd/config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	for _, test := range tests {
		correctPassword := "correct_password"

		correctHash, err := bcrypt.GenerateFromPassword([]byte(correctPassword), config.App.Security.Hash.Cost)
		if err != nil {
			t.Fatalf("Test %v, Hash() unexpected error: %v", test.name, err)
		}

		err = hasher.CompareHashAndPassword(correctHash, test.password)
		if err != nil && !test.wantErr {
			t.Fatalf("Test %v, Hash() unexpected error: %v", test.name, err)
		}
		if err == nil && test.wantErr {
			t.Fatalf("Test %v, Hash() unhandled error: %v", test.name, err)
		}
	}
}

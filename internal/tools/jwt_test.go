package tools

import (
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"errors"
	"testing"
)

func TestCreateToken(t *testing.T) {
	tests := []struct {
		name            string
		data            any
		wantErr         bool
		wantSpecificErr error
	}{
		{
			"successfull_generate_token",
			domain.DefaultUser{UserName: "test"},
			false,
			nil,
		},
	}

	if err := config.Init("../../cmd/config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	logger := testLogger{}
	tokenMaker := NewToken(logger)

	for _, test := range tests {
		_, err := tokenMaker.CreateToken(test.data)

		if err != nil {
			if !test.wantErr {
				t.Fatalf("Test %v, CreateToken() unexpected error: %v", test.name, err)
			}
			if test.wantSpecificErr != nil && !errors.Is(err, test.wantSpecificErr) {
				t.Fatalf("Test %v, CreateToken() unexpected error type: %v", test.name, err)
			}
		}

		if err == nil && test.wantErr {
			t.Fatalf("Test %v, CreateToken() unhandled error: %v", test.name, err)
		}

		t.Logf("Test %v, CreateToken() success", test.name)
	}
}

func TestValidateUserToken(t *testing.T) {
	tests := []struct {
		name            string
		valid           bool
		wantErr         bool
		wantSpecificErr error
	}{
		{
			"successfull_validate_token",
			true,
			false,
			nil,
		},

		{
			"error_invalid_token",
			false,
			true,
			domain.ErrInvalidToken,
		},
	}

	if err := config.Init("../../cmd/config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	logger := testLogger{}
	tokenMaker := NewToken(logger)

	invalidToken := "EyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzcwMjI0MzYsImlhdCI6MTc3NjkzNjAzNiwiaXNzIjoiYXBwIiwidXNlcm5hbWUiOiJ0ZXN0In0.HIRPan281i1plIzfaUl00uyvBsXM2u5FkCNWIp2nXVg"
	validToken, err := tokenMaker.CreateToken(domain.DefaultUser{UserName: "test"})
	if err != nil {
		t.Fatalf("Failed to create valid token: %v", err)
	}

	for _, test := range tests {
		var token string
		if test.valid {
			token = validToken
		} else {
			token = invalidToken
		}
		err = tokenMaker.ValidateUserToken(token)

		if err != nil {
			if !test.wantErr {
				t.Fatalf("Test %v, ValidateUserToken() unexpected error: %v", test.name, err)
			}
			if test.wantSpecificErr != nil && !errors.Is(err, test.wantSpecificErr) {
				t.Fatalf("Test %v, ValidateUserToken() unexpected error type: %v", test.name, err)
			}
		}

		if err == nil && test.wantErr {
			t.Fatalf("Test %v, ValidateUserToken() unhandled error: %v", test.name, err)
		}

		t.Logf("Test %v, ValidateUserToken() success", test.name)
	}
}

func TestParseUserTokenRaw(t *testing.T) {
	tests := []struct {
		name            string
		valid           bool
		wantErr         bool
		wantSpecificErr error
	}{
		{
			"successfull_parse_token",
			true,
			false,
			nil,
		},

		{
			"error_invalid_token",
			false,
			true,
			nil,
		},
	}

	if err := config.Init("../../cmd/config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	logger := testLogger{}
	tokenMaker := NewToken(logger)

	invalidToken := "EyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzcwMjI0MzYsImlhdCI6MTc3NjkzNjAzNiwiaXNzIjoiYXBwIiwidXNlcm5hbWUiOiJ0ZXN0In0.HIRPan281i1plIzfaUl00uyvBsXM2u5FkCNWIp2nXVg"
	validToken, err := tokenMaker.CreateToken(domain.DefaultUser{UserName: "test"})
	if err != nil {
		t.Fatalf("Failed to create valid token: %v", err)
	}

	for _, test := range tests {
		var token string
		if test.valid {
			token = validToken
		} else {
			token = invalidToken
		}
		_, err = tokenMaker.ParseUserTokenRaw(token)

		if err != nil {
			if !test.wantErr {
				t.Fatalf("Test %v, CreateToken() unexpected error: %v", test.name, err)
			}
			if test.wantSpecificErr != nil && !errors.Is(err, test.wantSpecificErr) {
				t.Fatalf("Test %v, CreateToken() unexpected error type: %v", test.name, err)
			}
		}

		if err == nil && test.wantErr {
			t.Fatalf("Test %v, CreateToken() unhandled error: %v", test.name, err)
		}

		t.Logf("Test %v, CreateToken() success", test.name)
	}
}

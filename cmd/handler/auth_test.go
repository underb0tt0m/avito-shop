package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/mocks"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
)

func TestAuth(t *testing.T) {
	if err := config.Init("../config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	method := http.MethodPost
	path := "/api/auth"
	logger := mocks.NewLogger(nil)

	tests := []struct {
		name           string
		body           any
		expectedStatus int
		expectedBody   any
		serviceSetup   func(s *mocks.MockServiceAuth)
	}{
		{
			"successful",
			dto.AuthRequest{
				Name:     "test",
				Password: "test",
			},
			http.StatusOK,
			dto.AuthResponse{Token: "test"},
			func(s *mocks.MockServiceAuth) {
				s.EXPECT().
					Auth(gomock.Any(), gomock.Any()).
					Return(dto.AuthResponse{Token: "test"}, nil)
			},
		},

		{
			"error_unprocessable_entity",
			`name`,
			domain.ErrUnprocessableEntity.Code,
			dto.ErrorResponse{Errors: domain.ErrUnprocessableEntity.Message},
			func(s *mocks.MockServiceAuth) {},
		},

		{
			"error_from_service",
			dto.AuthRequest{
				Name:     "test",
				Password: "test",
			},
			domain.ErrUnauthorized.Code,
			dto.ErrorResponse{Errors: domain.ErrUnauthorized.Message},
			func(s *mocks.MockServiceAuth) {
				s.EXPECT().
					Auth(gomock.Any(), gomock.Any()).
					Return(dto.AuthResponse{}, domain.ErrUnauthorized)
			},
		},
	}
	t.Logf("\n\n")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatalf("Failed to encode request body: %v", err)
			}
			req := httptest.NewRequest(method, path, &body)

			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serviceAuth := mocks.NewMockServiceAuth(ctrl)
			tt.serviceSetup(serviceAuth)

			router := chi.NewRouter()
			router.Route("/api", func(r chi.Router) {
				Auth(serviceAuth, r, logger)
			})

			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected %v, got %v", tt.expectedStatus, rr.Code)
			}

			switch expected := tt.expectedBody.(type) {
			case dto.AuthResponse:
				var got dto.AuthResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to unmarshal info response: %v; body=%s", err, rr.Body.String())
				}
				if !reflect.DeepEqual(got, expected) {
					t.Fatalf("expected body %+v, got %+v", expected, got)
				}

			case dto.ErrorResponse:
				var got dto.ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to unmarshal error response: %v; body=%s", err, rr.Body.String())
				}
				if !reflect.DeepEqual(got, expected) {
					t.Fatalf("expected body %+v, got %+v", expected, got)
				}

			default:
				t.Fatalf("unsupported expectedBody type %T", tt.expectedBody)
			}
			t.Logf("subtest %s passed, status: %v, body: %s", tt.name, rr.Code, rr.Body.String())
		})
	}
}

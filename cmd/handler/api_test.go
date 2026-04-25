package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/mocks"
	"avito-shop/internal/tools"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
)

func TestInfo(t *testing.T) {
	if err := config.Init("../config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	method := "GET"
	path := "/api/info"
	token := "Bearer test"
	logger := mocks.NewLogger(nil)
	tokenMaker := mocks.NewToken(nil, nil, nil)
	jsonCodec := tools.NewJSONCodec()

	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   any
		serviceSetup   func(s *mocks.MockServiceAPI)
	}{
		{
			"successful",
			http.StatusOK,
			dto.InfoResponse{
				Coins:     100,
				Inventory: []dto.Item{},
			},
			func(s *mocks.MockServiceAPI) {
				s.EXPECT().
					GetUserInfo(gomock.Any(), "test").
					Return(&dto.InfoResponse{
						Coins:     100,
						Inventory: []dto.Item{},
					}, nil)
			},
		},

		{
			"internal_server_error",
			domain.ErrInternalServerError.Code,
			dto.ErrorResponse{Errors: domain.ErrInternalServerError.Message},
			func(s *mocks.MockServiceAPI) {
				s.EXPECT().
					GetUserInfo(gomock.Any(), "test").
					Return(nil, domain.ErrInternalServerError)
			},
		},
	}
	t.Logf("\n\n")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(method, path, nil)
			req.Header.Set("Authorization", token)

			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serviceAPI := mocks.NewMockServiceAPI(ctrl)
			tt.serviceSetup(serviceAPI)

			router := chi.NewRouter()
			router.Route("/api", func(r chi.Router) {
				r.Use(mocks.Auth(logger, tokenMaker))
				Main(serviceAPI, r, logger, jsonCodec)
			})

			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected %v, got %v", tt.expectedStatus, rr.Code)
			}

			switch expected := tt.expectedBody.(type) {
			case dto.InfoResponse:
				var got dto.InfoResponse
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

func TestSendCoin(t *testing.T) {
	method := http.MethodPost
	path := "/api/sendCoin"
	token := "Bearer test"
	logger := mocks.NewLogger(nil)
	tokenMaker := mocks.NewToken(nil, nil, nil)
	jsonCodec := tools.NewJSONCodec()

	tests := []struct {
		name           string
		body           any
		serviceSetup   func(s *mocks.MockServiceAPI)
		expectedStatus int
		expectedBody   any
	}{
		{
			"successful",
			dto.SendCoinRequest{
				ToUser: "test",
				Amount: 100,
			},
			func(s *mocks.MockServiceAPI) {
				s.EXPECT().
					SendCoins(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(nil)
			},
			http.StatusOK,
			nil,
		},

		{
			"error_from_service",
			dto.SendCoinRequest{
				ToUser: "test",
				Amount: 100,
			},
			func(s *mocks.MockServiceAPI) {
				s.EXPECT().
					SendCoins(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(domain.ErrInsufficientFunds)
			},
			domain.ErrInsufficientFunds.Code,
			dto.ErrorResponse{Errors: domain.ErrInsufficientFunds.Message},
		},

		{
			"error_unprocessable_entity",
			`{"amount":"not a number"}`,
			func(s *mocks.MockServiceAPI) {},
			domain.ErrUnprocessableEntity.Code,
			dto.ErrorResponse{Errors: domain.ErrUnprocessableEntity.Message},
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
			req.Header.Set("Authorization", token)

			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serviceAPI := mocks.NewMockServiceAPI(ctrl)
			tt.serviceSetup(serviceAPI)

			router := chi.NewRouter()
			router.Route("/api", func(r chi.Router) {
				r.Use(mocks.Auth(logger, tokenMaker))
				Main(serviceAPI, r, logger, jsonCodec)
			})

			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected %v, got %v", tt.expectedStatus, rr.Code)
			}

			switch expected := tt.expectedBody.(type) {
			case nil:
				if rr.Body.Len() != 0 {
					t.Fatalf("expected nil body, got: %s", rr.Body.String())
				}
			case dto.ErrorResponse:
				var got dto.ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to unmarshal response: %v; body=%s", err, rr.Body.String())
				}
				if !reflect.DeepEqual(tt.expectedBody, got) {
					t.Fatalf("expected body %+v, got %+v", expected, got)
				}
			default:
				t.Fatalf("unsupported expectedBody type %T", tt.expectedBody)
			}
			t.Logf("subtest %s passed, status: %v, body: %s", tt.name, rr.Code, rr.Body.String())
		})
	}

}

func TestBuyItem(t *testing.T) {
	if err := config.Init("../config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	method := http.MethodPost
	basicPath := "/api/buy"
	token := "Bearer token"
	logger := mocks.NewLogger(nil)
	tokenMaker := mocks.NewToken(nil, nil, nil)
	jsonCodec := tools.NewJSONCodec()

	tests := []struct {
		name           string
		itemID         string
		body           any
		serviceSetup   func(s *mocks.MockServiceAPI)
		expectedStatus int
		expectedBody   any
	}{
		{
			"successful",
			"10",
			nil,
			func(s *mocks.MockServiceAPI) {
				s.EXPECT().
					BuyItem(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(nil)
			},
			http.StatusOK,
			nil,
		},

		{
			"error_bad_itemID",
			"test",
			nil,
			func(s *mocks.MockServiceAPI) {},
			domain.ErrBadRequest.Code,
			dto.ErrorResponse{Errors: domain.ErrBadRequest.Message},
		},

		{
			"error_from_service",
			"10",
			nil,
			func(s *mocks.MockServiceAPI) {
				s.EXPECT().
					BuyItem(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(domain.ErrInsufficientFunds)
			},
			domain.ErrInsufficientFunds.Code,
			dto.ErrorResponse{Errors: domain.ErrInsufficientFunds.Message},
		},
	}

	t.Logf("\n\n")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatalf("Failed to encode request body: %v", err)
			}

			fullPath := basicPath + "/" + tt.itemID
			if tt.itemID == "" {
				fullPath = basicPath + "/"
			}
			req := httptest.NewRequest(method, fullPath, &body)
			req.Header.Set("Authorization", token)

			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serviceAPI := mocks.NewMockServiceAPI(ctrl)
			tt.serviceSetup(serviceAPI)

			router := chi.NewRouter()
			router.Route("/api", func(r chi.Router) {
				r.Use(mocks.Auth(logger, tokenMaker))
				Main(serviceAPI, r, logger, jsonCodec)
			})

			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected %v, got %v", tt.expectedStatus, rr.Code)
			}

			switch expected := tt.expectedBody.(type) {
			case nil:
				if rr.Body.Len() != 0 {
					t.Fatalf("expected nil body, got: %s", rr.Body.String())
				}
			case dto.ErrorResponse:
				var got dto.ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to unmarshal response: %v; body=%s", err, rr.Body.String())
				}
				if !reflect.DeepEqual(tt.expectedBody, got) {
					t.Fatalf("expected body %+v, got %+v", expected, got)
				}
			default:
				t.Fatalf("unsupported expectedBody type %T", tt.expectedBody)
			}
			t.Logf("subtest %s passed, status: %v, body: %s", tt.name, rr.Code, rr.Body.String())
		})
	}
}

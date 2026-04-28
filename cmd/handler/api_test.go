package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/jsoncodec"
	"avito-shop/internal/logging"
	"avito-shop/internal/mocks"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMain_handleInfo(t *testing.T) {
	path := "/info"
	logger := logging.LoggerNoop{}
	jsonCodec := jsoncodec.NewJSONCodec("sonic")
	queryTimeout := time.Second

	type fields struct {
		logger       logging.Logger
		jsonCodec    jsoncodec.JSONCodec
		queryTimeout time.Duration
	}
	testFields := fields{
		logger:       logger,
		jsonCodec:    jsonCodec,
		queryTimeout: queryTimeout,
	}
	type mockSetups struct {
		service    func(m *mocks.ServiceAPI)
		tokenMaker func(m *mocks.TokenMaker)
	}
	tests := []struct {
		name           string
		fields         fields
		mockSetups     mockSetups
		withCtx        bool
		expectedStatus int
		expectedBody   dto.Response
	}{
		{
			name:   "successful",
			fields: testFields,
			mockSetups: mockSetups{
				service: func(m *mocks.ServiceAPI) {
					m.EXPECT().
						GetUserInfo(gomock.Any(), "successful").
						Return(&dto.InfoResponse{}, nil)
				},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx:        true,
			expectedStatus: http.StatusOK,
			expectedBody: dto.InfoResponse{
				Coins: 0,
				CoinHistory: dto.History{
					Received: nil,
					Sent:     nil,
				},
				Inventory: nil,
			},
		},

		{
			name:   "error_unauthorised",
			fields: testFields,
			mockSetups: mockSetups{
				service:    func(m *mocks.ServiceAPI) {},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrUnauthorized.Message},
		},

		{
			name:   "any_error_from_service",
			fields: testFields,
			mockSetups: mockSetups{
				service: func(m *mocks.ServiceAPI) {
					m.EXPECT().
						GetUserInfo(gomock.Any(), "any_error_from_service").
						Return(&dto.InfoResponse{}, errors.New("Some error"))
				},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx:        true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrInternalServerError.Message},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serviceMock := mocks.NewServiceAPI(ctrl)
			tokenMakerMock := mocks.NewTokenMaker(ctrl)

			req := httptest.NewRequest(http.MethodGet, path, nil)
			if tt.withCtx {
				ctx := context.WithValue(req.Context(), domain.UserContextKey, domain.DefaultUser{UserName: tt.name})
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			tt.mockSetups.service(serviceMock)
			tt.mockSetups.tokenMaker(tokenMakerMock)

			h := Main{
				service:      serviceMock,
				logger:       tt.fields.logger,
				jsonCodec:    tt.fields.jsonCodec,
				queryTimeout: tt.fields.queryTimeout,
				tokenMaker:   tokenMakerMock,
			}
			h.handleInfo(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			switch tt.expectedBody.(type) {
			case dto.InfoResponse:
				var body dto.InfoResponse
				if err := jsonCodec.Unmarshal(rr.Body.Bytes(), &body); err != nil {
					t.Fatalf(
						"failed to unmarshal body. expected: %+v, got: %s",
						tt.expectedBody,
						rr.Body.String(),
					)
				}
				assert.Equal(t, tt.expectedBody, body)
			case dto.ErrorResponse:
				var body dto.ErrorResponse
				if err := jsonCodec.Unmarshal(rr.Body.Bytes(), &body); err != nil {
					t.Fatalf(
						"unexpected body scheme. expected: %t, got: %s",
						tt.expectedBody,
						rr.Body.String(),
					)
				}
				assert.Equal(t, tt.expectedBody, body)
			default:
				t.Fatalf(
					"unexpected body scheme. expected: %t, got: %s",
					tt.expectedBody,
					rr.Body.String(),
				)
			}
		})
	}
}

func TestMain_handleSendCoin(t *testing.T) {
	path := "/sendCoin"

	logger := logging.LoggerNoop{}
	jsonCodec := jsoncodec.NewJSONCodec("sonic")
	queryTimeout := time.Second

	type fields struct {
		logger       logging.Logger
		jsonCodec    jsoncodec.JSONCodec
		queryTimeout time.Duration
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	testFields := fields{
		logger:       logger,
		jsonCodec:    jsonCodec,
		queryTimeout: queryTimeout,
	}
	type mockSetups struct {
		service    func(m *mocks.ServiceAPI)
		tokenMaker func(m *mocks.TokenMaker)
	}
	tests := []struct {
		name           string
		fields         fields
		mockSetups     mockSetups
		withCtx        bool
		requestBody    any
		expectedStatus int
		expectedBody   dto.Response
	}{
		{
			name:   "successful",
			fields: testFields,
			mockSetups: mockSetups{
				service: func(m *mocks.ServiceAPI) {
					m.EXPECT().
						SendCoins(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx: true,
			requestBody: dto.SendCoinRequest{
				ToUser: "test",
				Amount: 10,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},

		{
			name:   "error_unauthorized",
			fields: testFields,
			mockSetups: mockSetups{
				service:    func(m *mocks.ServiceAPI) {},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx: false,
			requestBody: dto.SendCoinRequest{
				ToUser: "test",
				Amount: 10,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrUnauthorized.Message},
		},

		{
			name:   "error_from_service",
			fields: testFields,
			mockSetups: mockSetups{
				service: func(m *mocks.ServiceAPI) {
					m.EXPECT().
						SendCoins(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(errors.New("Some error"))
				},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx: true,
			requestBody: dto.SendCoinRequest{
				ToUser: "test",
				Amount: 10,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrInternalServerError.Message},
		},

		{
			name:   "error_unprocessable_entity",
			fields: testFields,
			mockSetups: mockSetups{
				service:    func(m *mocks.ServiceAPI) {},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx:        true,
			requestBody:    `{bimbim: "bambam"}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrUnprocessableEntity.Message},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serviceMock := mocks.NewServiceAPI(ctrl)
			tt.mockSetups.service(serviceMock)
			tokenMakerMock := mocks.NewTokenMaker(ctrl)
			tt.mockSetups.tokenMaker(tokenMakerMock)

			rr := httptest.NewRecorder()

			var body bytes.Buffer
			_ = json.NewEncoder(&body).Encode(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, path, &body)
			if tt.withCtx {
				ctx := context.WithValue(req.Context(), domain.UserContextKey, domain.DefaultUser{UserName: tt.name})
				req = req.WithContext(ctx)
			}

			h := Main{
				service:      serviceMock,
				logger:       tt.fields.logger,
				jsonCodec:    tt.fields.jsonCodec,
				queryTimeout: tt.fields.queryTimeout,
				tokenMaker:   tokenMakerMock,
			}
			h.handleSendCoin(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			switch tt.expectedBody.(type) {
			case nil:
				if rr.Body.String() != "" {
					t.Fatalf(
						"unexpected body. expected: %+v, got: %s",
						tt.expectedBody,
						rr.Body.String(),
					)
				}
			case dto.ErrorResponse:
				var responseBody dto.ErrorResponse
				if err := jsonCodec.Unmarshal(rr.Body.Bytes(), &responseBody); err != nil {
					t.Fatalf(
						"unexpected body scheme. expected: %t, got: %s",
						tt.expectedBody,
						rr.Body.String(),
					)
				}
				assert.Equal(t, tt.expectedBody, responseBody)
			default:
				t.Fatalf(
					"unexpected body scheme. expected: %t, got: %s",
					tt.expectedBody,
					rr.Body.String(),
				)
			}
		})
	}
}

// "successful", "error_bad_itemID", "error_from_service",
func TestMain_handleBuyItem(t *testing.T) {
	path := "/buy/"

	logger := logging.LoggerNoop{}
	jsonCodec := jsoncodec.NewJSONCodec("sonic")
	queryTimeout := time.Second

	type fields struct {
		logger       logging.Logger
		jsonCodec    jsoncodec.JSONCodec
		queryTimeout time.Duration
	}
	testFields := fields{
		logger:       logger,
		jsonCodec:    jsonCodec,
		queryTimeout: queryTimeout,
	}
	type mockSetups struct {
		service    func(m *mocks.ServiceAPI)
		tokenMaker func(m *mocks.TokenMaker)
	}
	tests := []struct {
		name           string
		fields         fields
		mockSetups     mockSetups
		withCtx        bool
		itemID         string
		expectedStatus int
		expectedBody   dto.Response
	}{
		{
			name:   "successful",
			fields: testFields,
			mockSetups: mockSetups{
				service: func(m *mocks.ServiceAPI) {
					m.EXPECT().
						BuyItem(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx:        true,
			itemID:         "1",
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},

		{
			name:   "error_unauthorised",
			fields: testFields,
			mockSetups: mockSetups{
				service:    func(m *mocks.ServiceAPI) {},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx:        false,
			itemID:         "1",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrUnauthorized.Message},
		},

		{
			name:   "error_bad_itemID",
			fields: testFields,
			mockSetups: mockSetups{
				service:    func(m *mocks.ServiceAPI) {},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx:        true,
			itemID:         "hehmda",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrBadRequest.Message},
		},

		{
			name:   "error_from_service",
			fields: testFields,
			mockSetups: mockSetups{
				service: func(m *mocks.ServiceAPI) {
					m.EXPECT().
						BuyItem(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(errors.New("Some error"))
				},
				tokenMaker: func(m *mocks.TokenMaker) {},
			},
			withCtx:        true,
			itemID:         "1",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrInternalServerError.Message},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serviceMock := mocks.NewServiceAPI(ctrl)
			tt.mockSetups.service(serviceMock)
			tokenMakerMock := mocks.NewTokenMaker(ctrl)
			tt.mockSetups.tokenMaker(tokenMakerMock)

			rr := httptest.NewRecorder()

			testPath := path + tt.itemID
			req := httptest.NewRequest(http.MethodPost, testPath, nil)
			if tt.withCtx {
				ctx := context.WithValue(req.Context(), domain.UserContextKey, domain.DefaultUser{UserName: tt.name})
				req = req.WithContext(ctx)
			}
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
				URLParams: chi.RouteParams{
					Keys:   []string{"itemID"},
					Values: []string{tt.itemID},
				},
			})
			req = req.WithContext(ctx)

			h := Main{
				service:      serviceMock,
				logger:       tt.fields.logger,
				jsonCodec:    tt.fields.jsonCodec,
				queryTimeout: tt.fields.queryTimeout,
				tokenMaker:   tokenMakerMock,
			}
			h.handleBuyItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			switch tt.expectedBody.(type) {
			case nil:
				if rr.Body.String() != "" {
					t.Fatalf(
						"unexpected body. expected: %+v, got: %s",
						tt.expectedBody,
						rr.Body.String(),
					)
				}
			case dto.ErrorResponse:
				var responseBody dto.ErrorResponse
				if err := jsonCodec.Unmarshal(rr.Body.Bytes(), &responseBody); err != nil {
					t.Fatalf(
						"unexpected body scheme. expected: %t, got: %s",
						tt.expectedBody,
						rr.Body.String(),
					)
				}
				assert.Equal(t, tt.expectedBody, responseBody)
			default:
				t.Fatalf(
					"unexpected body scheme. expected: %t, got: %s",
					tt.expectedBody,
					rr.Body.String(),
				)
			}
		})
	}
}

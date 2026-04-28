package handler

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/jsoncodec"
	"avito-shop/internal/logging"
	"avito-shop/internal/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//"successful", "error_unprocessable_entity", "error_from_service",

func TestAuth_handleAuth(t *testing.T) {
	path := "/auth"

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
		service func(m *mocks.ServiceAuth)
	}
	tests := []struct {
		name           string
		fields         fields
		mockSetups     mockSetups
		requestBody    any
		expectedStatus int
		expectedBody   dto.Response
	}{
		{
			name:   "successful",
			fields: testFields,
			mockSetups: mockSetups{
				service: func(m *mocks.ServiceAuth) {
					m.EXPECT().
						Auth(gomock.Any(), gomock.Any()).
						Return(dto.AuthResponse{Token: "token"}, nil)
				},
			},
			requestBody: dto.AuthRequest{
				Name:     "test",
				Password: "test",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   dto.AuthResponse{Token: "token"},
		},

		{
			name:   "error_unprocessable_entity",
			fields: testFields,
			mockSetups: mockSetups{
				service: func(m *mocks.ServiceAuth) {},
			},
			requestBody:    `bimbim: "bambam"`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrUnprocessableEntity.Message},
		},

		{
			name:   "error_from_service",
			fields: testFields,
			mockSetups: mockSetups{
				service: func(m *mocks.ServiceAuth) {
					m.EXPECT().
						Auth(gomock.Any(), gomock.Any()).
						Return(dto.AuthResponse{}, errors.New("Some error"))
				},
			},
			requestBody: dto.AuthRequest{
				Name:     "test",
				Password: "test",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   dto.ErrorResponse{Errors: domain.ErrInternalServerError.Message},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serviceMock := mocks.NewServiceAuth(ctrl)
			tt.mockSetups.service(serviceMock)

			var body bytes.Buffer
			_ = json.NewEncoder(&body).Encode(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, path, &body)
			rr := httptest.NewRecorder()

			h := Auth{
				service:      serviceMock,
				logger:       tt.fields.logger,
				jsonCodec:    tt.fields.jsonCodec,
				queryTimeout: tt.fields.queryTimeout,
			}
			h.handleAuth(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			switch tt.expectedBody.(type) {
			case dto.AuthResponse:
				var body dto.AuthResponse
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

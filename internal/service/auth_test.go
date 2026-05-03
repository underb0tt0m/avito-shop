package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/mock/gomock"

	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/hasher"
	"avito-shop/internal/jwtmanager"
	"avito-shop/internal/logging"
	"avito-shop/internal/mocks"
	"avito-shop/internal/prometheus_metrics"
	"avito-shop/internal/storage"
)

func Test_auth_Auth(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	logger := logging.LoggerNoop{}

	hasherMock := mocks.NewHasher(ctrl)
	tokenMaker := mocks.NewTokenMaker(ctrl)
	storageMock := mocks.NewStorageAuth(ctrl)

	type fields struct {
		Storage    storage.Auth
		Logger     logging.Logger
		TokenMaker jwtmanager.TokenMaker
		Hasher     hasher.Hasher
	}
	type mockSetups struct {
		Storage    func(m *mocks.StorageAuth)
		TokenMaker func(m *mocks.TokenMaker)
		Hasher     func(m *mocks.Hasher)
	}
	type args struct {
		ctx  context.Context
		data dto.AuthRequest
	}
	tests := []struct {
		name            string
		fields          fields
		mockSetups      mockSetups
		args            args
		want            dto.AuthResponse
		wantErr         bool
		wantSpecificErr error
	}{
		{
			name: "success_existing_user_authenticated",
			fields: fields{
				Storage:    storageMock,
				Logger:     logger,
				TokenMaker: tokenMaker,
				Hasher:     hasherMock,
			},
			mockSetups: mockSetups{
				Storage: func(m *mocks.StorageAuth) {
					m.EXPECT().
						GetHashedUserPassword(ctx, "success_existing_user_authenticated").
						Return([]byte("success_existing_user_authenticated"), nil)
				},
				TokenMaker: func(m *mocks.TokenMaker) {
					m.EXPECT().
						CreateToken(domain.DefaultUser{UserName: "success_existing_user_authenticated"}).
						Return("success_existing_user_authenticated", nil)
				},
				Hasher: func(m *mocks.Hasher) {
					m.EXPECT().
						Hash("success_existing_user_authenticated", logger).
						Return([]byte("success_existing_user_authenticated"), nil)
					m.EXPECT().
						CompareHashAndPassword([]byte("success_existing_user_authenticated"), []byte("success_existing_user_authenticated")).
						Return(nil)
				},
			},
			args: args{
				ctx: ctx,
				data: dto.AuthRequest{
					Name:     "success_existing_user_authenticated",
					Password: "success_existing_user_authenticated",
				},
			},
			want:            dto.AuthResponse{Token: "success_existing_user_authenticated"},
			wantErr:         false,
			wantSpecificErr: nil,
		},

		{
			name: "error_wrong_password",
			fields: fields{
				Storage:    storageMock,
				Logger:     logger,
				TokenMaker: tokenMaker,
				Hasher:     hasherMock,
			},
			mockSetups: mockSetups{
				Storage: func(m *mocks.StorageAuth) {
					m.EXPECT().
						GetHashedUserPassword(ctx, "error_wrong_password").
						Return([]byte("error_wrong_password"), nil)
				},
				TokenMaker: func(_ *mocks.TokenMaker) {},
				Hasher: func(m *mocks.Hasher) {
					m.EXPECT().
						Hash("bimbimbambam", logger).
						Return([]byte("error_wrong_password"), nil)
					m.EXPECT().
						CompareHashAndPassword([]byte("error_wrong_password"), []byte("bimbimbambam")).
						Return(errors.New("Some error"))
				},
			},
			args: args{
				ctx: ctx,
				data: dto.AuthRequest{
					Name:     "error_wrong_password",
					Password: "bimbimbambam",
				},
			},
			want:            dto.AuthResponse{},
			wantErr:         true,
			wantSpecificErr: domain.ErrUnauthorized,
		},

		{
			name: "success_new_user_created",
			fields: fields{
				Storage:    storageMock,
				Logger:     logger,
				TokenMaker: tokenMaker,
				Hasher:     hasherMock,
			},
			mockSetups: mockSetups{
				Storage: func(m *mocks.StorageAuth) {
					m.EXPECT().
						GetHashedUserPassword(ctx, "success_new_user_created").
						Return(nil, pgx.ErrNoRows)
					m.EXPECT().
						CreateUser(ctx, domain.HashedUserData{
							Name:     "success_new_user_created",
							Password: []byte("success_new_user_created"),
						}).
						Return([]byte("success_new_user_created"), nil)
				},
				TokenMaker: func(m *mocks.TokenMaker) {
					m.EXPECT().
						CreateToken(domain.DefaultUser{UserName: "success_new_user_created"}).
						Return("success_new_user_created", nil)
				},
				Hasher: func(m *mocks.Hasher) {
					m.EXPECT().
						Hash("success_new_user_created", logger).
						Return([]byte("success_new_user_created"), nil)
					m.EXPECT().
						CompareHashAndPassword([]byte("success_new_user_created"), []byte("success_new_user_created")).
						Return(nil)
				},
			},
			args: args{
				ctx: ctx,
				data: dto.AuthRequest{
					Name:     "success_new_user_created",
					Password: "success_new_user_created",
				},
			},
			want:            dto.AuthResponse{Token: "success_new_user_created"},
			wantErr:         false,
			wantSpecificErr: nil,
		},

		{
			name: "error_database_unavailable_on_get",
			fields: fields{
				Storage:    storageMock,
				Logger:     logger,
				TokenMaker: tokenMaker,
				Hasher:     hasherMock,
			},
			mockSetups: mockSetups{
				Storage: func(m *mocks.StorageAuth) {
					m.EXPECT().
						GetHashedUserPassword(ctx, "error_database_unavailable_on_get").
						Return([]byte{}, errors.New("Some error"))
				},
				TokenMaker: func(_ *mocks.TokenMaker) {},
				Hasher: func(m *mocks.Hasher) {
					m.EXPECT().
						Hash("error_database_unavailable_on_get", logger).
						Return([]byte("error_database_unavailable_on_get"), nil)
				},
			},
			args: args{
				ctx: ctx,
				data: dto.AuthRequest{
					Name:     "error_database_unavailable_on_get",
					Password: "error_database_unavailable_on_get",
				},
			},
			want:            dto.AuthResponse{},
			wantErr:         true,
			wantSpecificErr: nil,
		},

		{
			name: "error_token_creation_failed",
			fields: fields{
				Storage:    storageMock,
				Logger:     logger,
				TokenMaker: tokenMaker,
				Hasher:     hasherMock,
			},
			mockSetups: mockSetups{
				Storage: func(m *mocks.StorageAuth) {
					m.EXPECT().
						GetHashedUserPassword(ctx, "error_token_creation_failed").
						Return([]byte("error_token_creation_failed"), nil)
				},
				TokenMaker: func(m *mocks.TokenMaker) {
					m.EXPECT().
						CreateToken(domain.DefaultUser{UserName: "error_token_creation_failed"}).
						Return("", errors.New("Some error"))
				},
				Hasher: func(m *mocks.Hasher) {
					m.EXPECT().
						Hash("error_token_creation_failed", logger).
						Return([]byte("error_token_creation_failed"), nil)
					m.EXPECT().
						CompareHashAndPassword([]byte("error_token_creation_failed"), []byte("error_token_creation_failed")).
						Return(nil)
				},
			},
			args: args{
				ctx: ctx,
				data: dto.AuthRequest{
					Name:     "error_token_creation_failed",
					Password: "error_token_creation_failed",
				},
			},
			want:            dto.AuthResponse{},
			wantErr:         true,
			wantSpecificErr: nil,
		},

		{
			name: "error_failed_to_hash_password",
			fields: fields{
				Storage:    storageMock,
				Logger:     logger,
				TokenMaker: tokenMaker,
				Hasher:     hasherMock,
			},
			mockSetups: mockSetups{
				Storage:    func(_ *mocks.StorageAuth) {},
				TokenMaker: func(_ *mocks.TokenMaker) {},
				Hasher: func(m *mocks.Hasher) {
					m.EXPECT().
						Hash("error_failed_to_hash_password", logger).
						Return([]byte{}, errors.New("Some error"))
				},
			},
			args: args{
				ctx: ctx,
				data: dto.AuthRequest{
					Name:     "error_failed_to_hash_password",
					Password: "error_failed_to_hash_password",
				},
			},
			want:            dto.AuthResponse{},
			wantErr:         true,
			wantSpecificErr: nil,
		},

		{
			name: "error_failed_to_create_user",
			fields: fields{
				Storage:    storageMock,
				Logger:     logger,
				TokenMaker: tokenMaker,
				Hasher:     hasherMock,
			},
			mockSetups: mockSetups{
				Storage: func(m *mocks.StorageAuth) {
					m.EXPECT().
						GetHashedUserPassword(ctx, "error_failed_to_create_user").
						Return([]byte{}, pgx.ErrNoRows)
					m.EXPECT().
						CreateUser(ctx, domain.HashedUserData{
							Name:     "error_failed_to_create_user",
							Password: []byte("error_failed_to_create_user"),
						}).
						Return([]byte{}, errors.New("Some error"))
				},
				TokenMaker: func(_ *mocks.TokenMaker) {},
				Hasher: func(m *mocks.Hasher) {
					m.EXPECT().
						Hash("error_failed_to_create_user", logger).
						Return([]byte("error_failed_to_create_user"), nil)
				},
			},
			args: args{
				ctx: ctx,
				data: dto.AuthRequest{
					Name:     "error_failed_to_create_user",
					Password: "error_failed_to_create_user",
				},
			},
			want:            dto.AuthResponse{},
			wantErr:         true,
			wantSpecificErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := auth{
				Storage:    tt.fields.Storage,
				Logger:     tt.fields.Logger,
				TokenMaker: tt.fields.TokenMaker,
				Hasher:     tt.fields.Hasher,
				Metrics:    prometheus_metrics.New(prometheus.NewRegistry()),
			}
			tt.mockSetups.Storage(storageMock)
			tt.mockSetups.Hasher(hasherMock)
			tt.mockSetups.TokenMaker(tokenMaker)

			got, err := s.Auth(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Auth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantSpecificErr != nil && !errors.Is(err, tt.wantSpecificErr) {
				t.Errorf("GetUserInfo() error = %v, wantSpecificErr %v", err, tt.wantSpecificErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Auth() got = %v, want %v", got, tt.want)
			}
		})
	}
}

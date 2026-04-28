package service

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/domain"
	"avito-shop/internal/logging"
	"avito-shop/internal/mocks"
	"avito-shop/internal/storage"
	"avito-shop/internal/storage/views"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

func Test_api_GetUserInfo(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	logger := logging.LoggerNoop{}

	storageMock := mocks.NewStorageAPI(ctrl)
	storageMock.EXPECT().
		GetUserInfo(ctx, "success_get_user_info").
		Return(1, []views.UserInventory{}, []views.UserTransaction{}, nil)
	storageMock.EXPECT().
		GetUserInfo(ctx, "error_missing_user").
		Return(0, nil, nil, pgx.ErrNoRows)
	storageMock.EXPECT().
		GetUserInfo(ctx, "error_database_unavailable_on_get").
		Return(0, nil, nil, errors.New("Some error"))

	type fields struct {
		Storage storage.API
		Logger  logging.Logger
	}
	type args struct {
		ctx      context.Context
		username string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		want            *dto.InfoResponse
		wantErr         bool
		wantSpecificErr error
	}{
		{
			name: "success_get_user_info",
			fields: fields{
				Storage: storageMock,
				Logger:  logger,
			},
			args: args{
				ctx:      ctx,
				username: "success_get_user_info",
			},
			want: &dto.InfoResponse{
				Coins:     1,
				Inventory: []dto.Item{},
				CoinHistory: dto.History{
					Received: []dto.ReceivedTransaction{},
					Sent:     []dto.SentTransaction{},
				},
			},
			wantErr:         false,
			wantSpecificErr: nil,
		},
		{
			name: "error_missing_user",
			fields: fields{
				Storage: storageMock,
				Logger:  logger,
			},
			args: args{
				ctx:      ctx,
				username: "error_missing_user",
			},
			want:            nil,
			wantErr:         true,
			wantSpecificErr: domain.ErrNotFound,
		},
		{
			name: "error_database_unavailable_on_get",
			fields: fields{
				Storage: storageMock,
				Logger:  logger,
			},
			args: args{
				ctx:      ctx,
				username: "error_database_unavailable_on_get",
			},
			want:            nil,
			wantErr:         true,
			wantSpecificErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := api{
				Storage: tt.fields.Storage,
				Logger:  tt.fields.Logger,
			}
			got, err := s.GetUserInfo(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantSpecificErr != nil && !errors.Is(err, tt.wantSpecificErr) {
				t.Errorf("GetUserInfo() error = %v, wantSpecificErr %v", err, tt.wantSpecificErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserInfo() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_api_SendCoins(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	logger := logging.LoggerNoop{}

	storageMock := mocks.NewStorageAPI(ctrl)
	storageMock.EXPECT().
		SendCoins(ctx, "success_send_coins", gomock.Any()).
		Return(nil)
	storageMock.EXPECT().
		SendCoins(ctx, "error_insufficient_funds", gomock.Any()).
		Return(domain.ErrInsufficientFunds)
	storageMock.EXPECT().
		SendCoins(ctx, "error_other_errors", gomock.Any()).
		Return(errors.New("Some error"))

	type fields struct {
		Storage storage.API
		Logger  logging.Logger
	}
	type args struct {
		ctx      context.Context
		fromUser string
		toUser   dto.SendCoinRequest
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErr         bool
		wantSpecificErr error
	}{
		{
			name: "success_send_coins",
			fields: fields{
				Storage: storageMock,
				Logger:  logger,
			},
			args: args{
				ctx:      ctx,
				fromUser: "success_send_coins",
				toUser: dto.SendCoinRequest{
					ToUser: "test",
					Amount: 10,
				},
			},
			wantErr:         false,
			wantSpecificErr: nil,
		},
		{
			name: "error_not_positive_amount",
			fields: fields{
				Storage: storageMock,
				Logger:  logger,
			},
			args: args{
				ctx:      ctx,
				fromUser: "test",
				toUser: dto.SendCoinRequest{
					ToUser: "test",
					Amount: -10,
				},
			},
			wantErr:         true,
			wantSpecificErr: domain.ErrBadRequest,
		},
		{
			name: "error_insufficient_funds",
			fields: fields{
				Storage: storageMock,
				Logger:  logger,
			},
			args: args{
				ctx:      ctx,
				fromUser: "error_insufficient_funds",
				toUser: dto.SendCoinRequest{
					ToUser: "test",
					Amount: 10,
				},
			},
			wantErr:         true,
			wantSpecificErr: domain.ErrInsufficientFunds,
		},
		{
			name: "error_other_errors",
			fields: fields{
				Storage: storageMock,
				Logger:  logger,
			},
			args: args{
				ctx:      ctx,
				fromUser: "error_other_errors",
				toUser: dto.SendCoinRequest{
					ToUser: "test",
					Amount: 10,
				},
			},
			wantErr:         true,
			wantSpecificErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := api{
				Storage: tt.fields.Storage,
				Logger:  tt.fields.Logger,
			}
			err := s.SendCoins(tt.args.ctx, tt.args.fromUser, tt.args.toUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendCoins() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantSpecificErr != nil && !errors.Is(err, tt.wantSpecificErr) {
				t.Errorf("GetUserInfo() error = %v, wantSpecificErr %v", err, tt.wantSpecificErr)
				return
			}
		})
	}
}

// "success_buy_item", "error_any",
func Test_api_BuyItem(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	logger := logging.LoggerNoop{}

	storageMock := mocks.NewStorageAPI(ctrl)
	storageMock.EXPECT().
		BuyItem(ctx, gomock.Any(), "success_buy_item").
		Return(nil)
	storageMock.EXPECT().
		BuyItem(ctx, gomock.Any(), "error_any").
		Return(errors.New("Some error"))

	type fields struct {
		Storage storage.API
		Logger  logging.Logger
	}
	type args struct {
		ctx    context.Context
		itemID int
		user   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success_buy_item",
			fields: fields{
				Storage: storageMock,
				Logger:  logger,
			},
			args: args{
				ctx:    ctx,
				itemID: 1,
				user:   "success_buy_item",
			},
			wantErr: false,
		},
		{
			name: "error_any",
			fields: fields{
				Storage: storageMock,
				Logger:  logger,
			},
			args: args{
				ctx:    ctx,
				itemID: 1,
				user:   "error_any",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := api{
				Storage: tt.fields.Storage,
				Logger:  tt.fields.Logger,
			}
			if err := s.BuyItem(tt.args.ctx, tt.args.itemID, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("BuyItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

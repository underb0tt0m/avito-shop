package service

import (
	"avito-shop/cmd/dto"
	"avito-shop/internal/config"
	"avito-shop/internal/domain"
	"avito-shop/internal/mocks"
	"avito-shop/internal/storage"
	"avito-shop/internal/storage/views"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestGetUserInfo(t *testing.T) {
	tests := []struct {
		name            string
		mockStorage     storage.API
		username        string
		expected        *dto.InfoResponse
		wantErr         bool
		wantSpecificErr error
	}{
		{
			"success_get_user_info",
			mocks.NewStorageAPI(
				func(ctx context.Context, username string) (
					int,
					[]views.UserInventory,
					[]views.UserTransaction,
					error,
				) {
					return 100,
						[]views.UserInventory{{"test", 1}},
						[]views.UserTransaction{
							{
								"friend",
								"test",
								10,
							},
							{
								"test",
								"friend",
								10,
							},
						},
						nil
				},
				nil,
				nil,
			),
			"test",
			&dto.InfoResponse{
				Coins: 100,
				Inventory: []dto.Item{
					{
						"test",
						1,
					},
				},
				CoinHistory: dto.History{
					Received: []dto.ReceivedTransaction{
						{
							"friend",
							10,
						},
					},
					Sent: []dto.SentTransaction{
						{
							"friend",
							10,
						},
					},
				},
			},
			false,
			nil,
		},

		{
			"success_user_with_empty_inventory",
			mocks.NewStorageAPI(
				func(ctx context.Context, username string) (
					int,
					[]views.UserInventory,
					[]views.UserTransaction, error) {
					return 100, nil, nil, nil
				},
				nil,
				nil,
			),
			"test",
			&dto.InfoResponse{
				Coins:     100,
				Inventory: []dto.Item{},
				CoinHistory: dto.History{
					Received: []dto.ReceivedTransaction{},
					Sent:     []dto.SentTransaction{},
				},
			},
			false,
			nil,
		},

		{
			"error_missing_user",
			mocks.NewStorageAPI(
				func(ctx context.Context, username string) (
					int,
					[]views.UserInventory,
					[]views.UserTransaction, error) {
					return 0, nil, nil, pgx.ErrNoRows
				},
				nil,
				nil,
			),
			"test",
			nil,
			true,
			domain.ErrNotFound,
		},

		{
			"error_database_unavailable_on_get",
			mocks.NewStorageAPI(
				func(ctx context.Context, username string) (int, []views.UserInventory, []views.UserTransaction, error) {
					return 0, nil, nil, errors.New("test")
				},
				nil,
				nil,
			),
			"test",
			nil,
			true,
			nil,
		},
	}

	if err := config.Init("../../cmd/config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	logger := mocks.NewLogger(nil)

	for _, test := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), config.App.Storage.QueryTimeout)
		TestAPIService := NewApi(
			test.mockStorage,
			logger,
		)

		result, err := TestAPIService.GetUserInfo(ctx, test.username)
		cancel()

		if err != nil {
			if !test.wantErr {
				t.Fatalf("Test %v, GetUserInfo() unexpected error: %v", test.name, err)
			}
			if test.wantSpecificErr != nil && !errors.Is(err, test.wantSpecificErr) {
				t.Errorf("Test %v, GetUserInfo() = %+v, want %+v", test.name, err, test.wantSpecificErr)
			}
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Test %v, GetUserInfo() = %+v, want %+v", test.name, result, test.expected)
		} else {
			t.Logf("Test %v, GetUserInfo() success: %+v", test.name, result)
		}
	}
}

func TestSendCoins(t *testing.T) {
	tests := []struct {
		name            string
		mockStorage     storage.API
		fromUser        string
		toUser          dto.SendCoinRequest
		wantErr         bool
		wantSpecificErr error
	}{
		{
			"success_send_coins",
			mocks.NewStorageAPI(nil, nil, nil),
			"test",
			dto.SendCoinRequest{Amount: 1},
			false,
			nil,
		},

		{
			"error_not_positive_amount",
			mocks.NewStorageAPI(nil, nil, nil),
			"test",
			dto.SendCoinRequest{},
			true,
			domain.ErrBadRequest,
		},

		{
			"error_insufficient_funds",
			mocks.NewStorageAPI(
				nil,
				func(ctx context.Context, fromUser string, transaction domain.SentTransaction) error {
					return domain.ErrInsufficientFunds
				},
				nil,
			),
			"test",
			dto.SendCoinRequest{Amount: 1},
			true,
			domain.ErrInsufficientFunds,
		},

		{
			"error_other_errors",
			mocks.NewStorageAPI(
				nil,
				func(ctx context.Context, fromUser string, transaction domain.SentTransaction) error {
					return errors.New("test")
				},
				nil,
			),
			"test",
			dto.SendCoinRequest{Amount: 1},
			true,
			nil,
		},
	}

	if err := config.Init("../../cmd/config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	logger := mocks.NewLogger(nil)

	for _, test := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), config.App.Storage.QueryTimeout)
		TestAPIService := NewApi(
			test.mockStorage,
			logger,
		)

		err := TestAPIService.SendCoins(ctx, test.fromUser, test.toUser)
		cancel()

		if err != nil {
			if !test.wantErr {
				t.Fatalf("Test %v, SendCoins() unexpected error: %v", test.name, err)
			}
			if test.wantSpecificErr != nil && !errors.Is(err, test.wantSpecificErr) {
				t.Errorf("Test %v, SendCoins() = %+v, want %+v", test.name, err, test.wantSpecificErr)
			}
		}
		if err == nil && !test.wantErr {
			t.Logf("Test %v, SendCoins() success", test.name)
		}
	}
}

func TestBuyItem(t *testing.T) {
	tests := []struct {
		name            string
		mockStorage     storage.API
		itemID          int
		user            string
		wantErr         bool
		wantSpecificErr error
	}{
		{
			"success_buy_item",
			mocks.NewStorageAPI(nil, nil, nil),
			1,
			"test",
			false,
			nil,
		},

		{
			"error_any",
			mocks.NewStorageAPI(
				nil,
				nil,
				func(ctx context.Context, itemID int, user string) error {
					return errors.New("test")
				},
			),
			1,
			"test",
			true,
			nil,
		},
	}

	if err := config.Init("../../cmd/config.yaml"); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	logger := mocks.NewLogger(nil)

	for _, test := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), config.App.Storage.QueryTimeout)
		TestAPIService := NewApi(
			test.mockStorage,
			logger,
		)

		err := TestAPIService.BuyItem(ctx, test.itemID, test.user)
		cancel()

		if err != nil {
			if !test.wantErr {
				t.Fatalf("Test %v, BuyItem() unexpected error: %v", test.name, err)
			}
			if test.wantSpecificErr != nil && !errors.Is(err, test.wantSpecificErr) {
				t.Errorf("Test %v, BuyItem() = %+v, want %+v", test.name, err, test.wantSpecificErr)
			}
		}
		if err == nil && !test.wantErr {
			t.Logf("Test %v, BuyItem() success", test.name)
		}
	}
}

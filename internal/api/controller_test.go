package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tebrizetayi/ledger_service/internal/api"
	"github.com/tebrizetayi/ledger_service/internal/storage"
	utils "github.com/tebrizetayi/ledger_service/internal/test_utils"
	"github.com/tebrizetayi/ledger_service/internal/transaction_manager"
)

var (
	GetUserBalanceTemplate            = "/users/%s/balance"
	GetUserTransactionHistoryTemplate = "/users/%s/history%s"
	AddTransactionTemplate            = "/users/%s/add"
)

func TestGetUserBalanceEndpoint(t *testing.T) {
	testCases := []struct {
		name               string
		userID             string
		expectedStatusCode int
		expectedBalance    float64
		mockBalance        float64
		mockError          error
	}{
		{
			name:               "Valid user ID",
			userID:             uuid.New().String(),
			expectedStatusCode: http.StatusOK,
			expectedBalance:    100.0,
			mockBalance:        100.0,
			mockError:          nil,
		},
		{
			name:               "Invalid user ID",
			userID:             "invalid-user-id",
			expectedStatusCode: http.StatusBadRequest,
			mockBalance:        100,
		},
	}

	for _, tc := range testCases {
		// Create a test environment
		testEnv, err := utils.CreateTestEnv()
		if err != nil {
			t.Fatalf("failed to create test env: %v", err)
		}
		defer testEnv.Cleanup()

		storageClient := storage.NewStorageClient(testEnv.DB)
		transactionManager := transaction_manager.NewTransactionManagerClient(storageClient)

		userId, _ := uuid.Parse(tc.userID)
		user := storage.User{
			ID:      userId,
			Balance: decimal.NewFromFloat(0),
		}

		err = storageClient.UserRepository.Add(testEnv.Context, user)
		if err != nil {
			t.Fatalf("failed to add user: %v", err)
		}

		_, err = transactionManager.AddTransaction(testEnv.Context, transaction_manager.Transaction{
			UserID:    userId,
			Amount:    decimal.NewFromFloat(tc.mockBalance),
			ID:        uuid.New(),
			CreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		})
		if err != nil {
			t.Fatalf("failed to add transaction: %v", err)
		}

		// Create the controller and the test request
		controller := api.NewController(transactionManager)
		newAPI := api.NewAPI(controller)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(GetUserBalanceTemplate, tc.userID), nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		rr := httptest.NewRecorder()
		newAPI.ServeHTTP(rr, req)

		// Check the response status code
		assert.Equal(t, tc.expectedStatusCode, rr.Code, fmt.Sprintf("expected status code %d, got %d", tc.expectedStatusCode, rr.Code))

		// If the status is OK, check the balance in the response
		if rr.Code == http.StatusOK {
			var response map[string]float64
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}
			assert.Equal(t, tc.expectedBalance, response["balance"])
		}
	}
}

func TestGetUserTransactionHistoryEndpoint(t *testing.T) {
	user := transaction_manager.User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}
	transactions := []transaction_manager.Transaction{
		{
			ID:             uuid.New(),
			UserID:         user.ID,
			Amount:         decimal.NewFromFloat(100),
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: uuid.New(),
		},
		{
			ID:             uuid.New(),
			UserID:         user.ID,
			Amount:         decimal.NewFromFloat(50),
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: uuid.New(),
		},
	}

	testCases := []struct {
		name                 string
		userID               string
		queryParams          string
		expectedStatusCode   int
		mockTransactions     []transaction_manager.Transaction
		expectedError        error
		expectedTransactions []transaction_manager.Transaction
	}{
		{
			name:               "Valid user ID",
			userID:             user.ID.String(),
			queryParams:        "?page=1&pageSize=10",
			expectedStatusCode: http.StatusOK,
			mockTransactions: []transaction_manager.Transaction{
				{
					ID:             transactions[0].ID,
					UserID:         transactions[0].UserID,
					Amount:         transactions[0].Amount,
					CreatedAt:      transactions[0].CreatedAt,
					IdempotencyKey: transactions[0].IdempotencyKey,
				},
				{
					ID:             transactions[1].ID,
					UserID:         transactions[1].UserID,
					Amount:         transactions[1].Amount,
					CreatedAt:      transactions[1].CreatedAt,
					IdempotencyKey: transactions[1].IdempotencyKey,
				},
			},
			expectedError: nil,
			expectedTransactions: []transaction_manager.Transaction{
				{
					ID:             transactions[0].ID,
					UserID:         transactions[0].UserID,
					Amount:         transactions[0].Amount,
					CreatedAt:      transactions[0].CreatedAt,
					IdempotencyKey: transactions[0].IdempotencyKey,
				},
				{
					ID:             transactions[1].ID,
					UserID:         transactions[1].UserID,
					Amount:         transactions[1].Amount,
					CreatedAt:      transactions[1].CreatedAt,
					IdempotencyKey: transactions[1].IdempotencyKey,
				},
			},
		},
		{
			name:               "Invalid user ID",
			userID:             "invalid-user-id",
			queryParams:        "?page=1&pageSize=10",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:                 "No transactions found",
			userID:               uuid.New().String(),
			queryParams:          "?page=1&pageSize=10",
			expectedStatusCode:   http.StatusOK,
			expectedTransactions: []transaction_manager.Transaction{},
			mockTransactions:     nil,
			expectedError:        nil,
		},
	}

	for _, tc := range testCases {
		// Create a test environment
		testEnv, err := utils.CreateTestEnv()
		if err != nil {
			t.Fatalf("failed to create test env: %v", err)
		}
		defer testEnv.Cleanup()

		storageClient := storage.NewStorageClient(testEnv.DB)
		transactionManager := transaction_manager.NewTransactionManagerClient(storageClient)

		userId, _ := uuid.Parse(tc.userID)
		user := storage.User{
			ID:      userId,
			Balance: decimal.NewFromFloat(0),
		}

		err = storageClient.UserRepository.Add(testEnv.Context, user)
		if err != nil {
			t.Fatalf("failed to add user: %v", err)
		}

		for i := range tc.mockTransactions {
			_, err = transactionManager.AddTransaction(testEnv.Context, transaction_manager.Transaction{
				UserID:         tc.mockTransactions[i].UserID,
				Amount:         tc.mockTransactions[i].Amount,
				ID:             tc.mockTransactions[i].ID,
				CreatedAt:      tc.mockTransactions[i].CreatedAt,
				IdempotencyKey: tc.mockTransactions[i].IdempotencyKey,
			})
			if err != nil {
				t.Fatalf("failed to add transaction: %v", err)
			}
		}

		// Create the controller and the test request
		controller := api.NewController(transactionManager)
		newAPI := api.NewAPI(controller)

		req, _ := http.NewRequest("GET", fmt.Sprintf(GetUserTransactionHistoryTemplate, tc.userID, tc.queryParams), nil)
		rr := httptest.NewRecorder()
		newAPI.ServeHTTP(rr, req)

		// Check the response status code
		assert.Equal(t, tc.expectedStatusCode, rr.Code, fmt.Sprintf("expected status code %d, got %d", tc.expectedStatusCode, rr.Code))

		// If the status is OK, check the transactions in the response
		if rr.Code == http.StatusOK {
			var transactions []transaction_manager.Transaction
			err = json.Unmarshal(rr.Body.Bytes(), &transactions)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			for i := range tc.expectedTransactions {
				found := false
				for _, expectedTransaction := range tc.expectedTransactions {
					if transactionsEqual(transactions[i], expectedTransaction) {
						found = true
						break
					}
				}
				assert.True(t, found, fmt.Sprintf("expected transaction %v, got %v", tc.expectedTransactions, transactions[i]))
			}
		}
	}
}

func TestAddTransaction(t *testing.T) {
	testUserID := uuid.New()
	testCases := []struct {
		name               string
		requestBody        []byte
		expectedStatusCode int
		mockError          error
	}{
		{
			name:               "Valid transaction",
			requestBody:        []byte(`{"user_id":"` + testUserID.String() + `", "amount":100}`),
			expectedStatusCode: http.StatusCreated,
			mockError:          nil,
		},
		{
			name:               "Invalid JSON",
			requestBody:        []byte(`{"user_id": , "amount":100}`),
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test environment
			testEnv, err := utils.CreateTestEnv()
			if err != nil {
				t.Fatalf("failed to create test env: %v", err)
			}
			defer testEnv.Cleanup()

			storageClient := storage.NewStorageClient(testEnv.DB)
			transactionManager := transaction_manager.NewTransactionManagerClient(storageClient)

			user := storage.User{
				ID:      testUserID,
				Balance: decimal.NewFromFloat(0),
			}

			err = storageClient.UserRepository.Add(testEnv.Context, user)
			if err != nil {
				t.Fatalf("failed to add user: %v", err)
			}

			// Create the controller and the test request
			controller := api.NewController(transactionManager)
			newAPI := api.NewAPI(controller)

			req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf(AddTransactionTemplate, testUserID.String()), bytes.NewBuffer(tc.requestBody))
			rr := httptest.NewRecorder()
			newAPI.ServeHTTP(rr, req)

			// Check the response status code
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			// If the status is StatusCreated, check the response message
			if rr.Code == http.StatusCreated {
				var response struct {
					Message string `json:"message"`
				}
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				assert.Equal(t, "Transaction successfully added", response.Message)

				transactions, err := transactionManager.GetUserTransactionHistory(testEnv.Context, testUserID, 1, 10)
				if err != nil {
					t.Fatalf("failed to get transactions: %v", err)
				}

				assert.Equal(t, 1, len(transactions))
				assert.Equal(t, testUserID, transactions[0].UserID)
				assert.Equal(t, transactions[0].Amount.Equal(decimal.NewFromFloat(100)), true)
			}
		})
	}
}

func transactionsEqual(a, b transaction_manager.Transaction) bool {
	return a.ID == b.ID &&
		a.Amount.Equal(b.Amount) &&
		a.UserID == b.UserID &&
		a.CreatedAt.Equal(b.CreatedAt) &&
		a.IdempotencyKey == b.IdempotencyKey
}

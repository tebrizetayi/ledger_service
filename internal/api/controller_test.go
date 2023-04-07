package api_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tebrizetayi/ledger_service/internal/api"
	"github.com/tebrizetayi/ledger_service/internal/storage"
	utils "github.com/tebrizetayi/ledger_service/internal/test_utils"
	"github.com/tebrizetayi/ledger_service/internal/transaction_manager"
)

var (
	GetUserBalanceTemplate            = "/users/%s/balance"
	GetUserTransactionHistoryTemplate = "/users/%s/transactions%s"
	AddTransactionTemplate            = "/users/%s/add"
)

func TestGetUserBalanceEndpoint(t *testing.T) {
	t.Skip("Skipping test for now")
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

		storageClient := storage.NewStorageClient(testEnv.DB)
		transactionManager := transaction_manager.NewTransactionManagerClient(storageClient)

		userId, _ := uuid.Parse(tc.userID)
		user := storage.User{
			ID:       userId,
			Username: "test",
		}

		err = storageClient.UserRepository.Add(testEnv.Context, user)
		if err != nil {
			t.Fatalf("failed to add user: %v", err)
		}

		_, err = transactionManager.AddTransaction(testEnv.Context, transaction_manager.Transaction{
			UserID:    userId,
			Amount:    tc.mockBalance,
			ID:        uuid.New(),
			CreatedAt: time.Now(),
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

func TestGetUserTransactionHistory_Endpoint(t *testing.T) {
	t.Skip("Skipping test for now")
	testUserID := uuid.New()
	testCases := []struct {
		name                 string
		userID               string
		queryParams          string
		expectedStatusCode   int
		mockTransactions     []transaction_manager.Transaction
		mockError            error
		expectedTransactions []transaction_manager.Transaction
	}{
		{
			name:               "Valid user ID",
			userID:             testUserID.String(),
			queryParams:        "?page=1&pageSize=10",
			expectedStatusCode: http.StatusOK,
			mockTransactions: []transaction_manager.Transaction{
				{
					ID:        uuid.New(),
					UserID:    testUserID,
					Amount:    100.0,
					CreatedAt: time.Now(),
				},
				{
					ID:        uuid.New(),
					UserID:    testUserID,
					Amount:    -50.0,
					CreatedAt: time.Now(),
				},
			},
			mockError:            nil,
			expectedTransactions: nil,
		},
		{
			name:               "Invalid user ID",
			userID:             "invalid-user-id",
			queryParams:        "?page=1&pageSize=10",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Error retrieving transaction history",
			userID:             uuid.New().String(),
			queryParams:        "?page=1&pageSize=10",
			expectedStatusCode: http.StatusInternalServerError,
			mockTransactions:   nil,
			mockError:          errors.New("Error retrieving transaction history"),
		},
	}

	for _, tc := range testCases {
		// Create a test environment
		testEnv, err := utils.CreateTestEnv()
		if err != nil {
			t.Fatalf("failed to create test env: %v", err)
		}
		storageClient := storage.NewStorageClient(testEnv.DB)
		transactionManager := transaction_manager.NewTransactionManagerClient(storageClient)

		userId, _ := uuid.Parse(tc.userID)
		user := storage.User{
			ID:       userId,
			Username: "test",
		}

		err = storageClient.UserRepository.Add(testEnv.Context, user)
		if err != nil {
			t.Fatalf("failed to add user: %v", err)
		}

		for i := range tc.mockTransactions {
			_, err = transactionManager.AddTransaction(testEnv.Context, transaction_manager.Transaction{
				UserID:    userId,
				Amount:    tc.mockTransactions[i].Amount,
				ID:        uuid.New(),
				CreatedAt: time.Now(),
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

		// Call the GetUserTransactionHistory method
		controller.GetUserTransactionHistory(rr, req)

		// Check the response status code
		assert.Equal(t, tc.expectedStatusCode, rr.Code)

		// If the status is OK, check the transactions in the response
		if rr.Code == http.StatusOK {
			var transactions []transaction_manager.Transaction
			err = json.Unmarshal(rr.Body.Bytes(), &transactions)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			assert.Equal(t, tc.mockTransactions, transactions)
		}
	}
}

/*

func TestAddTransaction(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        []byte
		expectedStatusCode int
		mockError          error
	}{
		{
			name:               "Valid transaction",
			requestBody:        []byte(`{"user_id":"` + uuid.New().String() + `", "amount":100}`),
			expectedStatusCode: http.StatusCreated,
			mockError:          nil,
		},
		{
			name:               "Invalid JSON",
			requestBody:        []byte(`{"user_id": , "amount":100}`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Error adding transaction",
			requestBody:        []byte(`{"user_id":"` + uuid.New().String() + `", "amount":100}`),
			expectedStatusCode: http.StatusInternalServerError,
			mockError:          errors.New("Error adding transaction"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test environment
			testEnv, err := utils.CreateTestEnv()
			if err != nil {
				t.Fatalf("failed to create test env: %v", err)
			}
		storageClient := storage.NewStorageClient(testEnv.DB)
		transactionManager := transaction_manager.NewTransactionManagerClient(storageClient)

			// Create the controller and the test request
			controller := api.NewController(transactionManager)
			newAPI := api.NewAPI(controller)

			req, _ := http.NewRequest("POST", fmt.Sprintf(AddTransactionTemplate), bytes.NewBuffer(tc.requestBody))
			rr := httptest.NewRecorder()
			newAPI.ServeHTTP(rr, req)

			// Check the response status code
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			// If the status is StatusCreated, check the response message
			if rr.Code != http.StatusCreated {
				t.Fatalf("expected status code %d, got %d", http.StatusCreated, rr.Code)
			}
			var response struct {
				Message string `json:"message"`
			}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}
			assert.Equal(t, "Transaction successfully added", response.Message)
		})
	}
}

*/

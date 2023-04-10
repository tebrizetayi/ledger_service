package storage

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	utils "github.com/tebrizetayi/ledger_service/internal/test_utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddTransaction_SingleTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)

	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId := user.ID

	expectedBalance := 100.00
	id := uuid.New()

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID:         user.ID,
			Amount:         decimal.NewFromFloat(100),
			ID:             id,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: uuid.New(),
		},
	})

	actualUser, err := userRepository.FindByID(testEnv.Context, userId)
	if err != nil {
		t.Fatalf("failed to get user balance: %v", err)
	}

	// Assert
	actualUserBalance, _ := actualUser.Balance.Float64()
	assert.Equal(t, expectedBalance, actualUserBalance, fmt.Sprintf("user balance should be %f actual %f", expectedBalance, actualUserBalance))
}

func TestAddTransaction_MultipleTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)
	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId := user.ID
	expectedBalance := 300.00

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID:         userId,
			Amount:         decimal.NewFromFloat(100),
			ID:             uuid.New(),
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: uuid.New(),
		},
		{
			UserID:         userId,
			Amount:         decimal.NewFromFloat(200),
			ID:             uuid.New(),
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: uuid.New(),
		},
	})

	actualUser, err := userRepository.FindByID(testEnv.Context, userId)
	if err != nil {
		t.Fatalf("failed to get user balance: %v", err)
	}

	// Assert
	actualUserBalance, _ := actualUser.Balance.Float64()
	assert.Equal(t, expectedBalance, actualUserBalance, fmt.Sprintf("user balance should be %f actual %f", expectedBalance, actualUserBalance))
}

func TestAddTransaction_AlreadyExistingTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)

	userRepository := NewUserRepository(testEnv.DB)
	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId := user.ID
	transactionID := uuid.New()
	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID:         userId,
			Amount:         decimal.NewFromFloat(100),
			ID:             transactionID,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: uuid.New(),
		},
	})

	_, err = transactionRepository.AddTransaction(testEnv.Context, Transaction{
		UserID: userId,
		Amount: decimal.NewFromFloat(100),
		ID:     transactionID,
	})
	assert.Error(t, err, "should return error when adding transaction with existing id")
}
func TestAddTransaction_SingleUser_Concurrent(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	userRepository := NewUserRepository(testEnv.DB)
	transactionRepository := NewTransactionRepository(testEnv.DB)

	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}
	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	numTransactions := 1000
	amountPerTransaction := decimal.NewFromFloat(1)
	expectedBalance := decimal.NewFromFloat(float64(numTransactions)).Mul(amountPerTransaction)

	// Act
	var wg sync.WaitGroup
	wg.Add(numTransactions)

	for i := 0; i < numTransactions; i++ {
		go func() {
			defer wg.Done()

			transaction := Transaction{
				ID:             uuid.New(),
				UserID:         user.ID,
				Amount:         amountPerTransaction,
				CreatedAt:      time.Now(),
				IdempotencyKey: uuid.New(),
			}

			_, err := transactionRepository.AddTransaction(testEnv.Context, transaction)
			if err != nil {
				t.Errorf("failed to add transaction: %v", err)
			}
		}()
	}

	wg.Wait()

	// Assert

	updatedUser, err := userRepository.FindByID(testEnv.Context, user.ID)
	assert.NoError(t, err, fmt.Sprintf("failed to get user: %v", err))
	assert.True(t, expectedBalance.Equal(updatedUser.Balance), fmt.Sprintf("expected balance %v actual %v", expectedBalance, updatedUser.Balance))
}
func TestAddTransaction_MultipleUsers_Concurrent(t *testing.T) {
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	userRepository := NewUserRepository(testEnv.DB)
	transactionRepository := NewTransactionRepository(testEnv.DB)

	numUsers := 100
	users := make([]User, numUsers)
	for i := 0; i < numUsers; i++ {
		user := User{
			ID:      uuid.New(),
			Balance: decimal.NewFromFloat(0),
		}
		err = userRepository.Add(testEnv.Context, user)
		if err != nil {
			t.Fatalf("failed to add user: %v", err)
		}
		users[i] = user
	}

	numTransactionsPerUser := 100
	amountPerTransaction := decimal.NewFromFloat(1)

	expectedBalance := decimal.NewFromFloat(float64(numTransactionsPerUser)).Mul(amountPerTransaction)

	var wg sync.WaitGroup
	wg.Add(numUsers * numTransactionsPerUser)

	for _, user := range users {
		for i := 0; i < numTransactionsPerUser; i++ {
			go func(userID uuid.UUID) {
				defer wg.Done()

				transaction := Transaction{
					ID:             uuid.New(),
					UserID:         userID,
					Amount:         amountPerTransaction,
					CreatedAt:      time.Now(),
					IdempotencyKey: uuid.New(),
				}

				_, err := transactionRepository.AddTransaction(testEnv.Context, transaction)
				if err != nil {
					t.Errorf("failed to add transaction: %v", err)
				}
			}(user.ID)
		}
	}

	wg.Wait()

	for _, user := range users {
		updatedUser, err := userRepository.FindByID(testEnv.Context, user.ID)
		assert.NoError(t, err, fmt.Sprintf("failed to get user: %v", err))
		assert.True(t, expectedBalance.Equal(updatedUser.Balance), fmt.Sprintf("expected balance %v actual %v", expectedBalance, updatedUser.Balance))
	}
}
func TestAddTransaction_InvalidUserID_Error(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)

	// Act
	_, err = transactionRepository.AddTransaction(testEnv.Context, Transaction{
		UserID:         uuid.New(),
		Amount:         decimal.NewFromFloat(100),
		ID:             uuid.New(),
		CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		IdempotencyKey: uuid.New(),
	})

	// Assert
	assert.Equal(t, ErrUserNotFound, err, fmt.Sprintf("expected error %v actual %v", ErrUserNotFound, err))
}
func TestGetUserTransactionHistory_MultipleTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)
	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId := user.ID

	transactionID1 := uuid.New()
	transactionID2 := uuid.New()

	idempotencyKey1 := uuid.New()
	idempotencyKey2 := uuid.New()
	expectedTransactions := []Transaction{
		{
			UserID:         userId,
			Amount:         decimal.NewFromFloat(100),
			ID:             transactionID1,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: idempotencyKey1,
		},
		{
			UserID:         userId,
			Amount:         decimal.NewFromFloat(300),
			ID:             transactionID2,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: idempotencyKey2,
		}}

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID:         userId,
			Amount:         decimal.NewFromFloat(100),
			ID:             transactionID1,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: idempotencyKey1,
		},
		{
			UserID:         userId,
			Amount:         decimal.NewFromFloat(300),
			ID:             transactionID2,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: idempotencyKey2,
		},
	})

	actualTransactions, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, userId, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	for i := range actualTransactions {
		found := false
		for _, expectedTransaction := range expectedTransactions {
			if transactionsEqual(actualTransactions[i], expectedTransaction) {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("expected transaction %v, got %v", expectedTransactions, actualTransactions[i]))
	}
}

func TestGetUserTransactionHistory_MultipleTransactionAndMultipleUser_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)
	user1 := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user1)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId1 := user1.ID

	user2 := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user2)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId2 := user2.ID

	transactionID11 := uuid.New()
	transactionID21 := uuid.New()
	IdempotentKey11 := uuid.New()
	IdempotentKey21 := uuid.New()

	expectedTransactions1 := []Transaction{
		{
			UserID:         userId1,
			Amount:         decimal.NewFromFloat(100),
			ID:             transactionID11,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: IdempotentKey11,
		},
		{
			UserID:         userId1,
			Amount:         decimal.NewFromFloat(300),
			ID:             transactionID21,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: IdempotentKey21,
		}}

	transactionID12 := uuid.New()
	transactionID22 := uuid.New()
	IdempotentKey12 := uuid.New()
	IdempotentKey22 := uuid.New()
	expectedTransactions2 := []Transaction{
		{
			UserID:         userId2,
			Amount:         decimal.NewFromFloat(800),
			ID:             transactionID12,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: IdempotentKey12,
		},
		{
			UserID:         userId2,
			Amount:         decimal.NewFromFloat(1000),
			ID:             transactionID22,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: IdempotentKey22,
		},
	}

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID:         userId1,
			Amount:         decimal.NewFromFloat(100),
			ID:             transactionID11,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: IdempotentKey11,
		},
		{
			UserID:         userId1,
			Amount:         decimal.NewFromFloat(300),
			ID:             transactionID21,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: IdempotentKey21,
		},
		{
			UserID:         userId2,
			Amount:         decimal.NewFromFloat(800),
			ID:             transactionID12,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: IdempotentKey12,
		},
		{
			UserID:         userId2,
			Amount:         decimal.NewFromFloat(1000),
			ID:             transactionID22,
			CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			IdempotencyKey: IdempotentKey22,
		},
	})

	actualTransactions1, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, userId1, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	actualTransactions2, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, userId2, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	for i := range actualTransactions1 {
		found := false
		for _, expectedTransaction := range expectedTransactions1 {
			if transactionsEqual(actualTransactions1[i], expectedTransaction) {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("expected transaction %v, got %v", expectedTransactions1, actualTransactions1[i]))
	}

	for i := range actualTransactions2 {
		found := false
		for _, expectedTransaction := range expectedTransactions2 {
			if transactionsEqual(actualTransactions2[i], expectedTransaction) {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("expected transaction %v, got %v", expectedTransactions2, actualTransactions2[i]))
	}
}

func TestGetUserTransactionHistory_MultiplePages(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)

	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}
	userRepository := NewUserRepository(testEnv.DB)
	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	numTransactions := 15
	pageSize := 5

	// Create transactions
	for i := 0; i < numTransactions; i++ {
		transaction := Transaction{
			ID:             uuid.New(),
			UserID:         user.ID,
			Amount:         decimal.NewFromFloat(float64(i + 1)),
			CreatedAt:      time.Now().Add(time.Duration(i) * time.Minute),
			IdempotencyKey: uuid.New(),
		}
		_, err = transactionRepository.AddTransaction(testEnv.Context, transaction)
		if err != nil {
			t.Fatalf("failed to add transaction: %v", err)
		}
	}

	// Act and Assert
	for pageNum := 1; pageNum <= (numTransactions / pageSize); pageNum++ {
		transactions, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, user.ID, pageNum, pageSize)
		assert.NoError(t, err)
		assert.Len(t, transactions, pageSize)

		// Check if transactions are in descending order by CreatedAt
		for i := 1; i < len(transactions); i++ {
			assert.True(t, transactions[i-1].CreatedAt.After(transactions[i].CreatedAt))
		}
	}
}

func TestGetUserTransactionHistory_EmptyResult_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)

	// Act
	transactions, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, uuid.New(), 1, 10)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, transactions)
}

func TestGetUserTransactionHistory_InvalidUserID_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)

	// Act
	transactions, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, uuid.New(), 1, 10)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, transactions)
}

func TestGetUserTransactionHistory_InvalidPageAndInvalidPageSize_Error(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)

	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}
	userRepository := NewUserRepository(testEnv.DB)
	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	numTransactions := 15
	// Create transactions
	for i := 0; i < numTransactions; i++ {
		transaction := Transaction{
			ID:             uuid.New(),
			UserID:         user.ID,
			Amount:         decimal.NewFromFloat(float64(i + 1)),
			CreatedAt:      time.Now().Add(time.Duration(i) * time.Minute),
			IdempotencyKey: uuid.New(),
		}
		_, err = transactionRepository.AddTransaction(testEnv.Context, transaction)
		if err != nil {
			t.Fatalf("failed to add transaction: %v", err)
		}
	}

	// Act and Assert
	transactions, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, user.ID, -1, -1)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(transactions))

}

func TestGetUserTransactionHistory_SingleTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)

	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId := user.ID
	expectedTransactions := []Transaction{
		{
			UserID:         userId,
			Amount:         decimal.NewFromFloat(100),
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			IdempotencyKey: uuid.New(),
		}}

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID:         userId,
			Amount:         decimal.NewFromFloat(100),
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			IdempotencyKey: uuid.New(),
		},
	})

	actualTransactions, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, userId, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	assert.Equal(t, len(expectedTransactions), len(actualTransactions), fmt.Sprintf("expected transaction count %v actual %v", len(expectedTransactions), len(actualTransactions)))
	assert.Equal(t, expectedTransactions[0].Amount, actualTransactions[0].Amount, fmt.Sprintf("expected transaction amount %v actual %v", expectedTransactions[0].Amount, actualTransactions[0].Amount))
	assert.Equal(t, expectedTransactions[0].UserID, actualTransactions[0].UserID, fmt.Sprintf("expected transaction user id, %d actual %d", expectedTransactions[0].UserID, actualTransactions[0].UserID))
}

func TestGetUserBalance_MultiTransactionAndSingleUser_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)
	user1 := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user1)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId1 := user1.ID
	expectedBalance1 := 400.00

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID:         userId1,
			Amount:         decimal.NewFromFloat(100),
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			IdempotencyKey: uuid.New(),
		},
		{
			UserID:         userId1,
			Amount:         decimal.NewFromFloat(300),
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			IdempotencyKey: uuid.New(),
		},
	})

	actualUser1, err := userRepository.FindByID(testEnv.Context, userId1)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	actualUserBalance1, _ := actualUser1.Balance.Float64()
	assert.Equal(t, expectedBalance1, actualUserBalance1, fmt.Sprintf("expected balance %v actual %v", expectedBalance1, actualUserBalance1))
}
func TestGetUserBalance_MultiTransactionAndMultiUser_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)
	user1 := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user1)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId1 := user1.ID

	user2 := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, user2)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId2 := user2.ID

	expectedBalance1 := 400.00
	expectedBalance2 := 1800.00

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID:         userId1,
			Amount:         decimal.NewFromFloat(100),
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			IdempotencyKey: uuid.New(),
		},
		{
			UserID:         userId1,
			Amount:         decimal.NewFromFloat(300),
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			IdempotencyKey: uuid.New(),
		},
		{
			UserID:         userId2,
			Amount:         decimal.NewFromFloat(800),
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			IdempotencyKey: uuid.New(),
		},
		{
			UserID:         userId2,
			Amount:         decimal.NewFromFloat(1000),
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			IdempotencyKey: uuid.New(),
		},
	})

	actualUser1, err := userRepository.FindByID(testEnv.Context, userId1)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	actualUser2, err := userRepository.FindByID(testEnv.Context, userId2)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	actualUserBalance1, _ := actualUser1.Balance.Float64()
	assert.Equal(t, expectedBalance1, actualUserBalance1, fmt.Sprintf("expected balance %v actual %v", expectedBalance1, actualUserBalance1))

	actualUserBalance2, _ := actualUser2.Balance.Float64()
	assert.Equal(t, expectedBalance2, actualUserBalance2, fmt.Sprintf("expected balance %v actual %v", expectedBalance2, actualUserBalance2))
}

func TestGetUserBalance_InvalidUserID_Error(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	userRepository := NewUserRepository(testEnv.DB)

	// Act
	_, err = userRepository.FindByID(testEnv.Context, uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func createTransactions(testEnv utils.TestEnv, transactionRepository *TransactionRepository, transactions []Transaction) error {
	for i := range transactions {
		_, err := transactionRepository.AddTransaction(testEnv.Context, transactions[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func transactionsEqual(a, b Transaction) bool {
	return a.ID == b.ID &&
		a.Amount.Equal(b.Amount) &&
		a.UserID == b.UserID &&
		a.CreatedAt.Equal(b.CreatedAt) &&
		a.IdempotencyKey == b.IdempotencyKey
}

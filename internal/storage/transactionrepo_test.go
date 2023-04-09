package storage

import (
	"fmt"
	"testing"

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
			UserID: user.ID,
			Amount: decimal.NewFromFloat(100),
			ID:     id,
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
			UserID: userId,
			Amount: decimal.NewFromFloat(100),
			ID:     uuid.New(),
		},
		{
			UserID: userId,
			Amount: decimal.NewFromFloat(200),
			ID:     uuid.New(),
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
			UserID: userId,
			Amount: decimal.NewFromFloat(100),
			ID:     transactionID,
		},
	})

	_, err = transactionRepository.AddTransaction(testEnv.Context, Transaction{
		UserID: userId,
		Amount: decimal.NewFromFloat(100),
		ID:     transactionID,
	})
	assert.Error(t, err, "should return error when adding transaction with existing id")
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
			UserID: userId,
			Amount: decimal.NewFromFloat(100),
			ID:     uuid.New(),
		}}

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID: userId,
			Amount: decimal.NewFromFloat(100),
			ID:     uuid.New(),
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

	expectedTransactions := []Transaction{
		{
			UserID: userId,
			Amount: decimal.NewFromFloat(100),
			ID:     transactionID1,
		},
		{
			UserID: userId,
			Amount: decimal.NewFromFloat(300),
			ID:     transactionID2,
		}}

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID: userId,
			Amount: decimal.NewFromFloat(100),
			ID:     transactionID1,
		},
		{
			UserID: userId,
			Amount: decimal.NewFromFloat(300),
			ID:     transactionID2,
		},
	})

	actualTransactions, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, userId, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	for i := range expectedTransactions {
		assert.Contains(t, expectedTransactions, Transaction{ID: actualTransactions[i].ID,
			UserID: actualTransactions[i].UserID,
			Amount: actualTransactions[i].Amount,
		}, fmt.Sprintf("expected transaction %v actual %v", expectedTransactions[i], actualTransactions[i]))
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
	expectedTransactions1 := []Transaction{
		{
			UserID: userId1,
			Amount: decimal.NewFromFloat(100),
			ID:     transactionID11,
		},
		{
			UserID: userId1,
			Amount: decimal.NewFromFloat(300),
			ID:     transactionID21,
		}}
	transactionID12 := uuid.New()
	transactionID22 := uuid.New()
	expectedTransactions2 := []Transaction{
		{
			UserID: userId2,
			Amount: decimal.NewFromFloat(800),
			ID:     transactionID12,
		},
		{
			UserID: userId2,
			Amount: decimal.NewFromFloat(1000),
			ID:     transactionID22,
		},
	}

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID: userId1,
			Amount: decimal.NewFromFloat(100),
			ID:     transactionID11,
		},
		{
			UserID: userId1,
			Amount: decimal.NewFromFloat(300),
			ID:     transactionID21,
		},
		{
			UserID: userId2,
			Amount: decimal.NewFromFloat(800),
			ID:     transactionID12,
		},
		{
			UserID: userId2,
			Amount: decimal.NewFromFloat(1000),
			ID:     transactionID22,
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
	for i := range expectedTransactions1 {
		assert.Contains(t, expectedTransactions1, Transaction{ID: actualTransactions1[i].ID, UserID: actualTransactions1[i].UserID, Amount: actualTransactions1[i].Amount}, fmt.Sprintf("expected transaction %v actual %v", expectedTransactions1[i], actualTransactions1[i]))
	}

	for i := range expectedTransactions2 {
		assert.Contains(t, expectedTransactions2, Transaction{ID: actualTransactions2[i].ID, UserID: actualTransactions2[i].UserID, Amount: actualTransactions2[i].Amount}, fmt.Sprintf("expected transaction %v actual %v", expectedTransactions2[i], actualTransactions2[i]))
	}
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
			UserID: userId1,
			Amount: decimal.NewFromFloat(100),
			ID:     uuid.New(),
		},
		{
			UserID: userId1,
			Amount: decimal.NewFromFloat(300),
			ID:     uuid.New(),
		},
		{
			UserID: userId2,
			Amount: decimal.NewFromFloat(800),
			ID:     uuid.New(),
		},
		{
			UserID: userId2,
			Amount: decimal.NewFromFloat(1000),
			ID:     uuid.New(),
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
		UserID: uuid.New(),
		Amount: decimal.NewFromFloat(100),
		ID:     uuid.New(),
	})

	// Assert
	assert.Equal(t, ErrUserNotFound, err, fmt.Sprintf("expected error %v actual %v", ErrUserNotFound, err))
}

func TestGetUserTransactionHistory_MultiplePages(t *testing.T) {
	t.Skip("not implemented")
}

func TestGetUserBalance_InvalidUserID_Error(t *testing.T) {
	t.Skip("not implemented")
}

func TestGetUserTransactionHistory_InvalidUserID_Error(t *testing.T) {
	t.Skip("not implemented")
}

func TestGetUserTransactionHistory_InvalidPage_Error(t *testing.T) {
	t.Skip("not implemented")
}

func TestGetUserTransactionHistory_InvalidPageSize_Error(t *testing.T) {
	t.Skip("not implemented")
}

func TestGetUserTransactionHistory_EmptyResult_Success(t *testing.T) {
	t.Skip("not implemented")
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

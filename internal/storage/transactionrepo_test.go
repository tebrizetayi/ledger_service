package storage

import (
	"fmt"
	"testing"

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

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)

	user := User{
		ID:       uuid.New(),
		Username: "test",
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
			Amount: 100,
			ID:     id,
		},
	})

	userBalance, err := transactionRepository.FindUserBalance(testEnv.Context, userId)
	if err != nil {
		t.Fatalf("failed to get user balance: %v", err)
	}

	// Assert
	assert.Equal(t, expectedBalance, userBalance, fmt.Sprintf("user balance should be %f", expectedBalance))
}

func TestAddTransaction_MultipleTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)
	user := User{
		ID:       uuid.New(),
		Username: "test",
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
			Amount: 100,
			ID:     uuid.New(),
		},
		{
			UserID: userId,
			Amount: 200,
			ID:     uuid.New(),
		},
	})

	userBalance, err := transactionRepository.FindUserBalance(testEnv.Context, userId)
	if err != nil {
		t.Fatalf("failed to get user balance: %v", err)
	}

	// Assert
	assert.Equal(t, userBalance, expectedBalance, fmt.Sprintf("user balance should be %f actual %f", expectedBalance, userBalance))
}

func TestAddTransaction_AlreadyExistingTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}

	transactionRepository := NewTransactionRepository(testEnv.DB)

	userRepository := NewUserRepository(testEnv.DB)
	user := User{
		ID:       uuid.New(),
		Username: "test",
	}

	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId := user.ID
	expectedBalance := 100.00
	transactionID := uuid.New()
	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID: userId,
			Amount: 100,
			ID:     transactionID,
		},
	})

	_, err = transactionRepository.AddTransaction(testEnv.Context, Transaction{
		UserID: userId,
		Amount: 100,
		ID:     transactionID,
	})
	if err != nil {
		t.Fatalf("failed to get user balance: %v", err)
	}

	userBalance, err := transactionRepository.FindUserBalance(testEnv.Context, userId)
	if err != nil {
		t.Fatalf("failed to get user balance: %v", err)
	}

	// Assert
	assert.Equal(t, userBalance, expectedBalance, fmt.Sprintf("user balance should be %f actual %f", expectedBalance, userBalance))
}

func TestGetUserTransactionHistory_SingleTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)

	user := User{
		ID:       uuid.New(),
		Username: "test",
	}

	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId := user.ID
	expectedTransactions := []Transaction{
		{
			UserID: userId,
			Amount: 100,
			ID:     uuid.New(),
		}}

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID: userId,
			Amount: 100,
			ID:     uuid.New(),
		},
	})

	actualTransactions, err := transactionRepository.GetUserTransactionHistory(testEnv.Context, userId, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	assert.Equal(t, len(expectedTransactions), len(actualTransactions), fmt.Sprintf("expected transaction count %v actual %v", len(expectedTransactions), len(actualTransactions)))
	assert.Equal(t, expectedTransactions[0].Amount, actualTransactions[0].Amount, fmt.Sprintf("expected transaction amount %f actual %f", expectedTransactions[0].Amount, actualTransactions[0].Amount))
	assert.Equal(t, expectedTransactions[0].UserID, actualTransactions[0].UserID, fmt.Sprintf("expected transaction user id, %d actual %d", expectedTransactions[0].UserID, actualTransactions[0].UserID))
}

func TestGetUserTransactionHistory_MultipleTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)
	user := User{
		ID:       uuid.New(),
		Username: "test",
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
			Amount: 100,
			ID:     transactionID1,
		},
		{
			UserID: userId,
			Amount: 300,
			ID:     transactionID2,
		}}

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID: userId,
			Amount: 100,
			ID:     transactionID1,
		},
		{
			UserID: userId,
			Amount: 300,
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

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)
	user1 := User{
		ID:       uuid.New(),
		Username: "test",
	}

	err = userRepository.Add(testEnv.Context, user1)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId1 := user1.ID

	user2 := User{
		ID:       uuid.New(),
		Username: "test",
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
			Amount: 100,
			ID:     transactionID11,
		},
		{
			UserID: userId1,
			Amount: 300,
			ID:     transactionID21,
		}}
	transactionID12 := uuid.New()
	transactionID22 := uuid.New()
	expectedTransactions2 := []Transaction{
		{
			UserID: userId2,
			Amount: 800,
			ID:     transactionID12,
		},
		{
			UserID: userId2,
			Amount: 1000,
			ID:     transactionID22,
		},
	}

	// Act
	createTransactions(testEnv, transactionRepository, []Transaction{
		{
			UserID: userId1,
			Amount: 100,
			ID:     transactionID11,
		},
		{
			UserID: userId1,
			Amount: 300,
			ID:     transactionID21,
		},
		{
			UserID: userId2,
			Amount: 800,
			ID:     transactionID12,
		},
		{
			UserID: userId2,
			Amount: 1000,
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

	transactionRepository := NewTransactionRepository(testEnv.DB)
	userRepository := NewUserRepository(testEnv.DB)
	user1 := User{
		ID:       uuid.New(),
		Username: "test",
	}

	err = userRepository.Add(testEnv.Context, user1)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}
	userId1 := user1.ID

	user2 := User{
		ID:       uuid.New(),
		Username: "test",
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
			Amount: 100,
			ID:     uuid.New(),
		},
		{
			UserID: userId1,
			Amount: 300,
			ID:     uuid.New(),
		},
		{
			UserID: userId2,
			Amount: 800,
			ID:     uuid.New(),
		},
		{
			UserID: userId2,
			Amount: 1000,
			ID:     uuid.New(),
		},
	})

	actualBalance1, err := transactionRepository.FindUserBalance(testEnv.Context, userId1)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	actualBalance2, err := transactionRepository.FindUserBalance(testEnv.Context, userId2)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	assert.Equal(t, expectedBalance1, actualBalance1)
	assert.Equal(t, expectedBalance2, actualBalance2)
}

func TestGetUserBalance_EmptyHistory_Error(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}

	transactionRepository := NewTransactionRepository(testEnv.DB)

	// Act
	actualBalance, err := transactionRepository.FindUserBalance(testEnv.Context, uuid.New())
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}
	// Assert
	assert.Zero(t, actualBalance)
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

func TestAddTransaction_InvalidUserID_Error(t *testing.T) {
	t.Skip("not implemented")
}

func TestGetUserTransactionHistory_MultiplePages(t *testing.T) {
	t.Skip("not implemented")
}

func TestAddTransaction_InvalidAmount_Error(t *testing.T) {
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

package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddTransaction_SingleTransaction_Success(t *testing.T) {
	// Assign
	ctx := context.Background()

	db, err := createTestDB(ctx)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	transactionRepository := NewTransactionRepository(db)

	userId := 1
	expectedBalance := 100.00

	// Act
	createTransactions(ctx, transactionRepository, []Transaction{
		{
			UserID: userId,
			Amount: 100,
		},
	})

	userBalance, err := transactionRepository.GetUserBalance(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get user balance: %v", err)
	}

	// Assert
	assert.Equal(t, userBalance, expectedBalance, fmt.Sprintf("user balance should be %f", expectedBalance))
}

func TestAddTransaction_MultipleTransaction_Success(t *testing.T) {
	// Assign
	ctx := context.Background()

	db, err := createTestDB(ctx)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	transactionRepository := NewTransactionRepository(db)

	userId := 1
	expectedBalance := 300.00

	// Act
	createTransactions(ctx, transactionRepository, []Transaction{
		{
			UserID: userId,
			Amount: 100,
		},
		{
			UserID: userId,
			Amount: 200,
		},
	})

	userBalance, err := transactionRepository.GetUserBalance(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get user balance: %v", err)
	}

	// Assert
	assert.Equal(t, userBalance, expectedBalance, fmt.Sprintf("user balance should be %f actual %f", expectedBalance, userBalance))
}

func TestGetUserTransactionHistory_SingleTransaction_Success(t *testing.T) {
	// Assign
	ctx := context.Background()

	db, err := createTestDB(ctx)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	transactionRepository := NewTransactionRepository(db)

	userId := 1
	expectedTransactions := []Transaction{
		{
			UserID: userId,
			Amount: 100,
		}}

	// Act
	createTransactions(ctx, transactionRepository, []Transaction{
		{
			UserID: userId,
			Amount: 100,
		},
	})

	actualTransactions, err := transactionRepository.GetUserTransactionHistory(ctx, 1, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	assert.Equal(t, len(expectedTransactions), len(actualTransactions), fmt.Sprintf("expected transaction count %v actual %v", len(expectedTransactions), len(actualTransactions)))
	assert.Equal(t, expectedTransactions[0].Amount, actualTransactions[0].Amount, fmt.Sprintf("expected transaction amount %f actual %f", expectedTransactions[0].Amount, actualTransactions[0].Amount))
	assert.Equal(t, expectedTransactions[0].UserID, actualTransactions[0].UserID, fmt.Sprintf("expected transaction user id %d actual %d", expectedTransactions[0].UserID, actualTransactions[0].UserID))
}

func TestGetUserTransactionHistory_MultipleTransaction_Success(t *testing.T) {
	// Assign
	ctx := context.Background()

	db, err := createTestDB(ctx)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	transactionRepository := NewTransactionRepository(db)

	userId := 1
	expectedTransactions := []Transaction{
		{
			UserID: userId,
			Amount: 100,
		},
		{
			UserID: userId,
			Amount: 300,
		}}

	// Act
	createTransactions(ctx, transactionRepository, []Transaction{
		{
			UserID: userId,
			Amount: 100,
		},
		{
			UserID: userId,
			Amount: 300,
		},
	})

	actualTransactions, err := transactionRepository.GetUserTransactionHistory(ctx, 1, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	for i := range expectedTransactions {
		assert.Contains(t, expectedTransactions, Transaction{UserID: actualTransactions[i].UserID, Amount: actualTransactions[i].Amount}, fmt.Sprintf("expected transaction %v actual %v", expectedTransactions[i], actualTransactions[i]))
	}
}

func TestGetUserTransactionHistory_MultipleTransactionAndMultipleUser_Success(t *testing.T) {
	// Assign
	ctx := context.Background()

	db, err := createTestDB(ctx)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	transactionRepository := NewTransactionRepository(db)

	userId1 := 1
	expectedTransactions1 := []Transaction{
		{
			UserID: userId1,
			Amount: 100,
		},
		{
			UserID: userId1,
			Amount: 300,
		}}

	userId2 := 2
	expectedTransactions2 := []Transaction{
		{
			UserID: userId2,
			Amount: 800,
		},
		{
			UserID: userId2,
			Amount: 1000,
		},
	}

	// Act
	createTransactions(ctx, transactionRepository, []Transaction{
		{
			UserID: userId1,
			Amount: 100,
		},
		{
			UserID: userId1,
			Amount: 300,
		},
		{
			UserID: userId2,
			Amount: 800,
		},
		{
			UserID: userId2,
			Amount: 1000,
		},
	})

	actualTransactions1, err := transactionRepository.GetUserTransactionHistory(ctx, 1, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	actualTransactions2, err := transactionRepository.GetUserTransactionHistory(ctx, 2, 1, 10)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	for i := range expectedTransactions1 {
		assert.Contains(t, expectedTransactions1, Transaction{UserID: actualTransactions1[i].UserID, Amount: actualTransactions1[i].Amount}, fmt.Sprintf("expected transaction %v actual %v", expectedTransactions1[i], actualTransactions1[i]))
	}

	for i := range expectedTransactions2 {
		assert.Contains(t, expectedTransactions2, Transaction{UserID: actualTransactions2[i].UserID, Amount: actualTransactions2[i].Amount}, fmt.Sprintf("expected transaction %v actual %v", expectedTransactions2[i], actualTransactions2[i]))
	}
}

func TestGetUserBalance_MultiTransactionAndMultiUser_Success(t *testing.T) {
	// Assign
	ctx := context.Background()

	db, err := createTestDB(ctx)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	transactionRepository := NewTransactionRepository(db)

	userId1 := 1
	expectedBalance1 := 400.00

	userId2 := 2
	expectedBalance2 := 1800.00

	// Act
	createTransactions(ctx, transactionRepository, []Transaction{
		{
			UserID: userId1,
			Amount: 100,
		},
		{
			UserID: userId1,
			Amount: 300,
		},
		{
			UserID: userId2,
			Amount: 800,
		},
		{
			UserID: userId2,
			Amount: 1000,
		},
	})

	actualBalance1, err := transactionRepository.GetUserBalance(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	actualBalance2, err := transactionRepository.GetUserBalance(ctx, 2)
	if err != nil {
		t.Fatalf("failed to get user transaction history: %v", err)
	}

	// Assert
	assert.Equal(t, expectedBalance1, actualBalance1)
	assert.Equal(t, expectedBalance2, actualBalance2)
}

func createTransactions(ctx context.Context, transactionRepository *TransactionRepository, transactions []Transaction) error {
	for i := range transactions {
		_, err := transactionRepository.AddTransaction(ctx, transactions[i])
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

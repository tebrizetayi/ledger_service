package transaction_manager

import (
	"testing"

	"github.com/tebrizetayi/ledger_service/internal/storage"
	utils "github.com/tebrizetayi/ledger_service/internal/test_utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddTransaction_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create test env: %v", err)
	}
	defer utils.CleanUpTestEnv(&testEnv)

	transactionManager := NewTransactionManagerClient(testEnv.DB)

	user := storage.User{
		ID:       uuid.New(),
		Username: "test",
	}
	err = transactionManager.storageClient.UserRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	// Act
	transaction, err := transactionManager.AddTransaction(testEnv.Context, Transaction{
		ID:     uuid.New(),
		Amount: 100,
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	// Assert
	assert.Equal(t, 100.0, transaction.Amount)
	assert.Equal(t, user.ID, transaction.UserID)
}

func TestAddTransaction_NotValidAmount(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create test env: %v", err)
	}
	defer utils.CleanUpTestEnv(&testEnv)

	transactionManager := NewTransactionManagerClient(testEnv.DB)

	user := storage.User{
		ID:       uuid.New(),
		Username: "test",
	}
	err = transactionManager.storageClient.UserRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	// Act
	_, err = transactionManager.AddTransaction(testEnv.Context, Transaction{
		ID:     uuid.New(),
		Amount: 0,
		UserID: user.ID,
	})

	// Assert
	assert.Equal(t, ErrInvalidTransaction, err)
}

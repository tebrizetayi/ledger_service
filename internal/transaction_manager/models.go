package transaction_manager

import (
	"database/sql"
	"errors"
	"ledger_service/internal/storage"
	"time"

	"context"

	"github.com/google/uuid"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
)

type Storage interface {
	storage.StorageClient
}

type TransactionManager struct {
	StorageClient storage.StorageClient
}

type Transaction struct {
	ID        uuid.UUID `json:"id"`
	Amount    float64   `json:"amount"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{
		StorageClient: storage.NewStorageClient(db),
	}
}

func (tm *TransactionManager) AddTransaction(ctx context.Context, transactionEntity Transaction) (Transaction, error) {
	if !tm.ValidateTransaction(ctx, transactionEntity) {
		return Transaction{}, ErrInvalidTransaction
	}

	transactionEntity.CreatedAt = time.Now()

	_, err := tm.StorageClient.TransactionRepository.AddTransaction(ctx, storage.Transaction{
		ID:        transactionEntity.ID,
		Amount:    transactionEntity.Amount,
		UserID:    transactionEntity.UserID,
		CreatedAt: transactionEntity.CreatedAt,
	})
	if err != nil {
		return Transaction{}, err
	}

	return transactionEntity, nil
}

func (tm *TransactionManager) ValidateTransaction(ctx context.Context, transaction Transaction) bool {
	// Validate the transaction
	_, err := uuid.Parse(transaction.UserID.String())
	if err != nil {
		return false
	}

	if transaction.Amount <= 0 {
		return false
	}

	return true
}

func (tm *TransactionManager) GetUserBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	// Validate the user
	_, err := tm.StorageClient.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return 0, err
	}

	balance, err := tm.StorageClient.TransactionRepository.FindUserBalance(ctx, userID)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

func (tm *TransactionManager) GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, page int, pageSize int) ([]Transaction, error) {
	// Validate the user
	_, err := tm.StorageClient.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	transactionResult, err := tm.StorageClient.TransactionRepository.GetUserTransactionHistory(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	transactions := make([]Transaction, len(transactionResult))
	for i, transaction := range transactionResult {
		transactions[i] = Transaction{
			ID:        transaction.ID,
			Amount:    transaction.Amount,
			UserID:    transaction.UserID,
			CreatedAt: transaction.CreatedAt,
		}
	}
	return transactions, nil
}

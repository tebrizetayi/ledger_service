package transaction_manager

import (
	"database/sql"
	"errors"
	"time"

	"github.com/tebrizetayi/ledger_service/internal/storage"

	"context"

	"github.com/google/uuid"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
)

type Storage interface {
	storage.StorageClient
}

type TransactionManagerClient struct {
	storageClient storage.StorageClient
}

type Transaction struct {
	ID        uuid.UUID `json:"id"`
	Amount    float64   `json:"amount"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewTransactionManagerClient(db *sql.DB) *TransactionManagerClient {
	return &TransactionManagerClient{
		storageClient: storage.NewStorageClient(db),
	}
}

func (tm *TransactionManagerClient) AddTransaction(ctx context.Context, transactionEntity Transaction) (Transaction, error) {
	if !tm.ValidateTransaction(ctx, transactionEntity) {
		return Transaction{}, ErrInvalidTransaction
	}

	transactionEntity.CreatedAt = time.Now()

	_, err := tm.storageClient.TransactionRepository.AddTransaction(ctx, storage.Transaction{
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

func (tm *TransactionManagerClient) ValidateTransaction(ctx context.Context, transaction Transaction) bool {
	// Validate the transaction
	return transaction.Amount <= 0.00
}

func (tm *TransactionManagerClient) GetUserBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	// Validate the user
	_, err := tm.storageClient.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return 0, err
	}

	balance, err := tm.storageClient.TransactionRepository.FindUserBalance(ctx, userID)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

func (tm *TransactionManagerClient) GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, page int, pageSize int) ([]Transaction, error) {
	// Validate the user
	_, err := tm.storageClient.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	transactionResult, err := tm.storageClient.TransactionRepository.GetUserTransactionHistory(ctx, userID, page, pageSize)
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

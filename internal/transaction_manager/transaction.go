package transaction_manager

import (
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tebrizetayi/ledger_service/internal/storage"

	"context"

	"github.com/google/uuid"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
)

func NewTransactionManagerClient(storage storage.StorageClient) *TransactionManagerClient {
	return &TransactionManagerClient{
		storageClient: storage,
	}
}

func (tm *TransactionManagerClient) AddTransaction(ctx context.Context, transactionEntity Transaction) (Transaction, error) {
	if !tm.ValidateTransaction(ctx, transactionEntity) {
		return Transaction{}, ErrInvalidTransaction
	}

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
	return transaction.Amount.IsPositive()
}

func (tm *TransactionManagerClient) GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	user, err := tm.storageClient.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return decimal.NewFromFloat(0), err
	}

	return user.Balance, nil
}

func (tm *TransactionManagerClient) GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, page int, pageSize int) ([]Transaction, error) {
	// Validate the user
	_, err := tm.storageClient.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return []Transaction{}, err
	}

	transactionResult, err := tm.storageClient.TransactionRepository.GetUserTransactionHistory(ctx, userID, page, pageSize)
	if err != nil {
		return []Transaction{}, err
	}

	transactions := []Transaction{}
	for _, transaction := range transactionResult {
		transactions = append(transactions, Transaction{
			ID:        transaction.ID,
			Amount:    transaction.Amount,
			UserID:    transaction.UserID,
			CreatedAt: transaction.CreatedAt,
		})
	}
	return transactions, nil
}

func (tm *TransactionManagerClient) IsUserValid(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := tm.storageClient.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	return user.ID != uuid.Nil, nil
}

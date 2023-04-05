package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Amount    float64
	CreatedAt time.Time
}

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (t *TransactionRepository) FindTransactionByID(ctx context.Context, transactionID uuid.UUID) (Transaction, error) {
	var transaction Transaction
	err := t.db.QueryRowContext(ctx, `SELECT id, user_id, amount, created_at FROM transactions WHERE id = $1`, transactionID).Scan(&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.CreatedAt)
	return transaction, err
}

func (t *TransactionRepository) AddTransaction(ctx context.Context, transaction Transaction) (Transaction, error) {
	var existingTransaction Transaction
	err := t.db.QueryRowContext(ctx, `SELECT id, user_id, amount FROM transactions WHERE id = $1`, transaction.ID).Scan(&existingTransaction.ID, &existingTransaction.UserID, &existingTransaction.Amount)

	if err == nil {
		// Transaction with the same UUID already exists, return the existing one
		return existingTransaction, nil
	}

	if err != sql.ErrNoRows {
		// Error occurred while querying the database
		return Transaction{}, err
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return Transaction{}, err
	}

	var transactionID uuid.UUID
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `INSERT INTO transactions (id,user_id, amount, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id, created_at`, transaction.ID, transaction.UserID, transaction.Amount).Scan(&transactionID, &createdAt)
	if err != nil {
		tx.Rollback()
		return Transaction{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Transaction{}, err
	}

	return Transaction{
		ID:        transactionID,
		UserID:    transaction.UserID,
		Amount:    transaction.Amount,
		CreatedAt: createdAt,
	}, nil
}

func (t *TransactionRepository) FindUserBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	var balance float64
	err := t.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount), 0)  FROM transactions WHERE user_id = $1`, userID).Scan(&balance)
	return balance, err
}

func (t *TransactionRepository) GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, page int, pageSize int) ([]Transaction, error) {
	rows, err := t.db.QueryContext(ctx, `SELECT id,user_id, amount, created_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		var transaction Transaction
		err = rows.Scan(&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

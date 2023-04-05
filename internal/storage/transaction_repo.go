package storage

import (
	"context"
	"database/sql"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (t *TransactionRepository) AddTransaction(ctx context.Context, transaction Transaction) (Transaction, error) {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return Transaction{}, err
	}

	var transactionID int
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `INSERT INTO transactions (user_id, amount, created_at) VALUES ($1, $2, NOW()) RETURNING id, created_at`, transaction.UserID, transaction.Amount).Scan(&transactionID, &createdAt)
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

func (t *TransactionRepository) GetUserBalance(ctx context.Context, userID int64) (float64, error) {
	var balance float64
	err := t.db.QueryRowContext(ctx, `SELECT sum(amount) FROM transactions WHERE user_id = $1`, userID).Scan(&balance)
	return balance, err
}

func (t *TransactionRepository) GetUserTransactionHistory(ctx context.Context, userID int64, page int, pageSize int) ([]Transaction, error) {
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

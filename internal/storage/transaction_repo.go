package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Amount    decimal.Decimal
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
	tx, err := t.db.Begin()
	if err != nil {
		return Transaction{}, err
	}

	// Lock the user row using SELECT FOR UPDATE
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = $1 FOR UPDATE", transaction.UserID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return Transaction{}, ErrUserNotFound
	}

	if err != nil {
		tx.Rollback()
		return Transaction{}, err
	}

	// Insert the transaction
	err = tx.QueryRowContext(ctx, `INSERT INTO transactions (id, user_id, amount, created_at) VALUES ($1, $2, $3, $4) RETURNING id, created_at`, transaction.ID, transaction.UserID, transaction.Amount, transaction.CreatedAt).Scan(&transaction.ID, &transaction.CreatedAt)
	if err != nil {
		tx.Rollback()
		return Transaction{}, err
	}

	// Update the user's balance
	newBalance := decimal.NewFromFloat(currentBalance).Add(transaction.Amount)
	_, err = tx.Exec("UPDATE users SET balance = $1 WHERE id = $2", newBalance, transaction.UserID)
	if err != nil {
		tx.Rollback()
		return Transaction{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Transaction{}, err
	}

	return Transaction{
		ID:        transaction.ID,
		UserID:    transaction.UserID,
		Amount:    transaction.Amount,
		CreatedAt: transaction.CreatedAt,
	}, nil
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

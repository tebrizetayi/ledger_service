package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Amount         decimal.Decimal
	CreatedAt      time.Time
	IdempotencyKey uuid.UUID
}

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (t *TransactionRepository) FindTransactionByID(ctx context.Context, transactionID uuid.UUID) (Transaction, error) {
	var transaction Transaction
	err := t.db.QueryRowContext(ctx, `SELECT id, user_id, amount, created_at,idempotency_key FROM transactions WHERE id = $1`, transactionID).
		Scan(&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.CreatedAt,
			&transaction.IdempotencyKey)

	return transaction, err
}

func (t *TransactionRepository) AddTransaction(ctx context.Context, transaction Transaction) (Transaction, error) {
	// Begin a new transaction
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return Transaction{}, err
	}

	// Lock the user row using SELECT FOR UPDATE
	var currentBalance decimal.Decimal
	err = tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = $1 FOR UPDATE", transaction.UserID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return Transaction{}, ErrUserNotFound
	}

	if err != nil {
		tx.Rollback()
		return Transaction{}, err
	}

	// Insert the transaction
	err = tx.QueryRowContext(ctx, `INSERT INTO transactions (id, user_id, amount, created_at,idempotency_key) VALUES ($1, $2, $3, $4,$5) RETURNING id, created_at`,
		transaction.ID,
		transaction.UserID,
		transaction.Amount,
		transaction.CreatedAt,
		transaction.IdempotencyKey).
		Scan(&transaction.ID,
			&transaction.CreatedAt)
	if err != nil {
		tx.Rollback()
		return Transaction{}, err
	}

	// Update the user's balance
	newBalance := currentBalance.Add(transaction.Amount)
	_, err = tx.ExecContext(ctx, "UPDATE users SET balance = $1 WHERE id = $2", newBalance, transaction.UserID)
	if err != nil {
		tx.Rollback()
		return Transaction{}, err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return Transaction{}, err
	}

	return Transaction{
		ID:             transaction.ID,
		UserID:         transaction.UserID,
		Amount:         transaction.Amount,
		CreatedAt:      transaction.CreatedAt,
		IdempotencyKey: transaction.IdempotencyKey,
	}, nil
}

func (t *TransactionRepository) GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, page int, pageSize int) ([]Transaction, error) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	rows, err := t.db.QueryContext(ctx, `SELECT id,user_id, amount, created_at,idempotency_key FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}

	transactions := []Transaction{}
	for rows.Next() {
		var transaction Transaction
		err = rows.Scan(&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.CreatedAt,
			&transaction.IdempotencyKey,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (t *TransactionRepository) FindTransactionByIdempotencyKey(ctx context.Context, idempotencyKey uuid.UUID) (Transaction, error) {
	var transaction Transaction
	err := t.db.QueryRowContext(ctx, `SELECT id, user_id, amount, created_at,idempotency_key FROM transactions WHERE idempotency_key = $1`, idempotencyKey).
		Scan(&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.CreatedAt,
			&transaction.IdempotencyKey)
	return transaction, err
}

package storage

import (
	"time"
)

type Transaction struct {
	ID        int
	UserID    int
	Amount    float64
	CreatedAt time.Time
}

type Storage interface {
	TransactionRepository
}

/*
type TransactionRepository interface {
	AddTransaction(userID int64, amount float64) (Transaction, error)
	GetUserBalance(userID int64) (float64, error)
	GetUserTransactionHistory(userID int64, page int, pageSize int) ([]Transaction, error)
}
*/

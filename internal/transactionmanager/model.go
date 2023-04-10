package transactionmanager

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tebrizetayi/ledgerservice/internal/storage"
)

type Storage interface {
	storage.StorageClient
}

type TransactionManagerClient struct {
	storageClient storage.StorageClient
}

type Transaction struct {
	ID             uuid.UUID       `json:"id"`
	Amount         decimal.Decimal `json:"amount"`
	UserID         uuid.UUID       `json:"user_id"`
	CreatedAt      time.Time       `json:"created_at"`
	IdempotencyKey uuid.UUID       `json:"idempotency_key"` // Add idempotency key to the transaction struct
}

type User struct {
	ID      uuid.UUID
	Balance decimal.Decimal
}

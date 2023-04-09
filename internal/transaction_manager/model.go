package transaction_manager

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tebrizetayi/ledger_service/internal/storage"
)

type Storage interface {
	storage.StorageClient
}

type TransactionManagerClient struct {
	storageClient storage.StorageClient
}

type Transaction struct {
	ID        uuid.UUID       `json:"id"`
	Amount    decimal.Decimal `json:"amount"`
	UserID    uuid.UUID       `json:"user_id"`
	CreatedAt time.Time       `json:"created_at"`
}
type User struct {
	ID      uuid.UUID
	Balance decimal.Decimal
}

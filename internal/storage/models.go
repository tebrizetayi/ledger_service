package storage

import (
	"database/sql"
)

type StorageClient struct {
	TransactionRepository *TransactionRepository
	UserRepository        *UserRepository
}

func NewStorageClient(db *sql.DB) StorageClient {
	return StorageClient{
		TransactionRepository: NewTransactionRepository(db),
		UserRepository:        NewUserRepository(db),
	}
}

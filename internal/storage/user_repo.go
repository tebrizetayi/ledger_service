package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID      uuid.UUID
	Balance decimal.Decimal
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

var ErrUserNotFound = errors.New("user not found")

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, "SELECT id, balance FROM users WHERE id = $1", id).Scan(&user.ID, &user.Balance)

	if err == sql.ErrNoRows {
		return User{}, ErrUserNotFound
	}

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *UserRepository) Add(ctx context.Context, u User) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO users (id, balance) VALUES ($1, $2)", u.ID, u.Balance)
	if err != nil {
		return err
	}
	return nil
}

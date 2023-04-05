package storage

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Username string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) CheckIf(ctx context.Context, username string) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, "SELECT id, username FROM users WHERE username = $1", username).Scan(&u.ID, &u.Username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, "SELECT id, username FROM users WHERE id = $1", id).Scan(&u.ID, &u.Username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Add(ctx context.Context, u User) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO users (id, username) VALUES ($1, $2)", u.ID, u.Username)
	if err != nil {
		return err
	}
	return nil
}

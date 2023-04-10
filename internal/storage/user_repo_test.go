package storage

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	utils "github.com/tebrizetayi/ledgerservice/internal/test_utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddUser_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	userRepository := NewUserRepository(testEnv.DB)

	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	// Act
	err = userRepository.Add(testEnv.Context, user)

	// Assert
	assert.NoError(t, err)
}

func TestAddUser_ExistingUser_Error(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	userRepository := NewUserRepository(testEnv.DB)

	exitingUser := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(0),
	}

	err = userRepository.Add(testEnv.Context, exitingUser)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	// Act
	err = userRepository.Add(testEnv.Context, exitingUser)
	assert.Error(t, err)
}

func TestAddUser_MultipleUsers_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	userRepository := NewUserRepository(testEnv.DB)

	users := []User{}
	for i := 0; i < 3; i++ {
		user := User{
			ID:      uuid.New(),
			Balance: decimal.NewFromFloat(100),
		}
		users = append(users, user)
	}

	// Act
	for _, user := range users {
		err = userRepository.Add(testEnv.Context, user)
		assert.NoError(t, err)
	}
}

func TestFindByID_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	userRepository := NewUserRepository(testEnv.DB)

	userID := uuid.New()
	user := User{
		ID:      userID,
		Balance: decimal.NewFromFloat(100),
	}

	err = userRepository.Add(testEnv.Context, user)
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	// Act
	foundUser, err := userRepository.FindByID(testEnv.Context, userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Balance, foundUser.Balance)
}

func TestFindByID_NotFound_Error(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	userRepository := NewUserRepository(testEnv.DB)

	// Act
	_, err = userRepository.FindByID(testEnv.Context, uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestAddUser_ContextCancel_Error(t *testing.T) {
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	ctx, cancel := context.WithCancel(testEnv.Context)
	testEnv.Context = ctx

	userRepository := NewUserRepository(testEnv.DB)

	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(10),
	}

	// Act
	cancel()
	err = userRepository.Add(testEnv.Context, user)
	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestAddUser_ContextTimeout_Error(t *testing.T) {
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer testEnv.Cleanup()

	ctx, cancel := context.WithTimeout(testEnv.Context, 1*time.Millisecond)
	defer cancel()
	testEnv.Context = ctx

	userRepository := NewUserRepository(testEnv.DB)

	user := User{
		ID:      uuid.New(),
		Balance: decimal.NewFromFloat(10),
	}

	// Act
	time.Sleep(6 * time.Millisecond)
	err = userRepository.Add(testEnv.Context, user)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

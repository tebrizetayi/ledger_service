package storage

import (
	"context"
	"testing"

	utils "github.com/tebrizetayi/ledger_service/internal/test_utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddUser_Success(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer utils.CleanUpTestEnv(&testEnv)

	userRepository := NewUserRepository(testEnv.DB)

	user := User{
		ID:       uuid.New(),
		Username: "test",
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
	defer utils.CleanUpTestEnv(&testEnv)

	userRepository := NewUserRepository(testEnv.DB)

	exitingUser := User{
		ID:       uuid.New(),
		Username: "test",
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
	defer utils.CleanUpTestEnv(&testEnv)

	userRepository := NewUserRepository(testEnv.DB)

	users := []User{}
	for i := 0; i < 100; i++ {
		user := User{
			ID:       uuid.New(),
			Username: "test",
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
	defer utils.CleanUpTestEnv(&testEnv)

	userRepository := NewUserRepository(testEnv.DB)

	userID := uuid.New()
	user := User{
		ID:       userID,
		Username: "test",
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
	assert.Equal(t, user.Username, foundUser.Username)
}

func TestFindByID_NotFound_Error(t *testing.T) {
	// Assign
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer utils.CleanUpTestEnv(&testEnv)

	userRepository := NewUserRepository(testEnv.DB)

	// Act
	_, err = userRepository.FindByID(testEnv.Context, uuid.New())

	// Assert
	assert.Error(t, err)
}

func TestAddUser_ContextCancel_Error(t *testing.T) {
	testEnv, err := utils.CreateTestEnv()
	if err != nil {
		t.Fatalf("failed to create env: %v", err)
	}
	defer utils.CleanUpTestEnv(&testEnv)

	ctx, cancel := context.WithCancel(testEnv.Context)
	testEnv.Context = ctx

	userRepository := NewUserRepository(testEnv.DB)

	user := User{
		ID:       uuid.New(),
		Username: "test",
	}

	// Act
	cancel()
	err = userRepository.Add(testEnv.Context, user)
	assert.Error(t, err)
}

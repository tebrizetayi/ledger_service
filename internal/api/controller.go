package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tebrizetayi/ledgerservice/internal/transactionmanager"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TransactionManager interface {
	AddTransaction(ctx context.Context, transaction transactionmanager.Transaction) (transactionmanager.Transaction, error)
	GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, page int, pageSize int) ([]transactionmanager.Transaction, error)
}

type Controller struct {
	transactionmanager TransactionManager
}

func NewController(tm TransactionManager) Controller {
	return Controller{
		transactionmanager: tm,
	}
}

type AddTransactionRequest struct {
	Amount         float64   `json:"amount"`
	IdempotencyKey uuid.UUID `json:"idempotency_key"`
}

func (c *Controller) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["uid"])
	if err != nil {
		httpError(w, fmt.Sprintf("Invalid user ID %s", err), http.StatusBadRequest)
		return
	}

	balance, err := c.transactionmanager.GetUserBalance(ctx, userID)
	if err != nil {
		httpError(w, fmt.Sprintf("Error retrieving user balance %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]decimal.Decimal{
		"balance": balance,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (c *Controller) AddTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["uid"])
	if err != nil {
		httpError(w, fmt.Sprintf("Invalid user ID %s", err), http.StatusBadRequest)
		return
	}

	var addTransactionRequest AddTransactionRequest
	if err := decodeJSON(r, &addTransactionRequest); err != nil {
		httpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	transaction := transactionmanager.Transaction{
		UserID:         userID,
		Amount:         decimal.NewFromFloat(addTransactionRequest.Amount),
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		IdempotencyKey: addTransactionRequest.IdempotencyKey,
	}

	if _, err := c.transactionmanager.AddTransaction(ctx, transaction); err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Message string `json:"message"`
	}{
		Message: "Transaction successfully added",
	}
	respondWithJSON(w, http.StatusCreated, response)
}

func (c *Controller) GetUserTransactionHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["uid"])
	if err != nil {
		httpError(w, fmt.Sprintf("Invalid user ID %v", err), http.StatusBadRequest)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	transactions, err := c.transactionmanager.GetUserTransactionHistory(ctx, userID, page, pageSize)
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, transactions)
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func httpError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}{
		Error:   http.StatusText(statusCode),
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}

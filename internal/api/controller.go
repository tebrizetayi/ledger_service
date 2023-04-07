package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tebrizetayi/ledger_service/internal/transaction_manager"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TransactionManager interface {
	AddTransaction(ctx context.Context, transaction transaction_manager.Transaction) (transaction_manager.Transaction, error)
	GetUserBalance(ctx context.Context, userID uuid.UUID) (float64, error)
	GetUserTransactionHistory(ctx context.Context, userID uuid.UUID, page int, pageSize int) ([]transaction_manager.Transaction, error)
}

type Controller struct {
	transaction_manager TransactionManager
}

func NewController(tm TransactionManager) Controller {
	return Controller{
		transaction_manager: tm,
	}
}

func (c *Controller) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["uid"])
	if err != nil {
		httpError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	balance, err := c.transaction_manager.GetUserBalance(ctx, userID)
	if err != nil {
		httpError(w, "Error retrieving user balance", http.StatusInternalServerError)
		return
	}

	response := map[string]float64{
		"balance": balance,
	}
	respondWithJSON(w, http.StatusOK, response)
}

type AddTransactionRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Amount float64   `json:"amount"`
}

func (c *Controller) AddTransaction(w http.ResponseWriter, r *http.Request) {
	var addTransactionRequest AddTransactionRequest
	if err := decodeJSON(r, &addTransactionRequest); err != nil {
		httpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	transaction := transaction_manager.Transaction{
		UserID: addTransactionRequest.UserID,
		Amount: addTransactionRequest.Amount,
	}

	ctx := r.Context()
	if _, err := c.transaction_manager.AddTransaction(ctx, transaction); err != nil {
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
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["uid"])
	if err != nil {
		httpError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 10
	}

	ctx := r.Context()
	transactions, err := c.transaction_manager.GetUserTransactionHistory(ctx, userID, page, pageSize)
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

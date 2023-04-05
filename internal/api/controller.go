package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"ledger_service/internal/transaction_manager"

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
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	balance, err := c.transaction_manager.GetUserBalance(ctx, userID)
	if err != nil {
		http.Error(w, "Error retrieving user balance", http.StatusInternalServerError)
		return
	}

	response := map[string]float64{
		"balance": balance,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// Decode the request body
type AddTransactionRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Amount float64   `json:"amount"`
}

func (c *Controller) AddTransaction(w http.ResponseWriter, r *http.Request) {
	addTransactionRequest := AddTransactionRequest{}
	if err := json.NewDecoder(r.Body).Decode(&addTransactionRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the transaction
	transaction := transaction_manager.Transaction{
		UserID: addTransactionRequest.UserID,
		Amount: addTransactionRequest.Amount,
	}

	// Add the transaction
	ctx := r.Context()
	if _, err := c.transaction_manager.AddTransaction(ctx, transaction); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a success response
	w.WriteHeader(http.StatusCreated)

	response := struct {
		Message string `json:"message"`
	}{
		Message: "Transaction successfully added",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) GetUserTransactionHistory(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the URL path
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["uid"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse pagination query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 10
	}

	// Get the transaction history
	ctx := r.Context()
	transactions, err := c.transaction_manager.GetUserTransactionHistory(ctx, userID, page, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the transaction history as a JSON response
	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

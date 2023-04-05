package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

const (
	addTransaction = "/users/{uid}/add"
	getUserBalance = "/users/{uid}/balance"
	userHistory    = "/users/{uid}/history"
)

func NewAPI(apiController Controller) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc(addTransaction, apiController.AddTransaction).Methods(http.MethodPost)
	router.HandleFunc(getUserBalance, apiController.GetUserBalance).Methods(http.MethodGet)
	router.HandleFunc(userHistory, apiController.GetUserTransactionHistory).Methods(http.MethodGet)
	return router
}

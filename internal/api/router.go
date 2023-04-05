package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

const (
	FindCommonManagerEndPoint = "/findcommonmanager"
)

func NewAPI(apiController Controller) http.Handler {
	router := mux.NewRouter()
	return router
}

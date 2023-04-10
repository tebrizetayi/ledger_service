package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
)

const (
	addTransaction = "/users/{uid}/add"
	getUserBalance = "/users/{uid}/balance"
	userHistory    = "/users/{uid}/history"
)

func NewAPI(apiController Controller) http.Handler {
	router := mux.NewRouter()
	router.Handle(addTransaction, rateLimitedMiddleware(apiController.AddTransaction)).Methods(http.MethodPost)
	router.Handle(getUserBalance, rateLimitedMiddleware(apiController.GetUserBalance)).Methods(http.MethodGet)
	router.Handle(userHistory, rateLimitedMiddleware(apiController.GetUserTransactionHistory)).Methods(http.MethodGet)

	return router
}

func rateLimitedMiddleware(next http.HandlerFunc) http.Handler {
	// Create a new rate limiter with a limit of 1 requests per second.
	rateLimiter := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: 0})

	// You can also customize the limiter with additional configurations.
	// For example, set a custom message and content type for the HTTP response when a client exceeds the rate limit.
	rateLimiter.SetMessage("You have reached the maximum request limit.")
	rateLimiter.SetMessageContentType("text/plain; charset=utf-8")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpError := tollbooth.LimitByRequest(rateLimiter, w, r)
		if httpError != nil {
			// Return a 429 status code if the rate limit is exceeded.
			w.Header().Add("Content-Type", rateLimiter.GetMessageContentType())
			w.WriteHeader(httpError.StatusCode)
			w.Write([]byte(httpError.Message))
			return
		}

		// Call the next handler if the rate limit is not exceeded.
		next.ServeHTTP(w, r)
	})
}

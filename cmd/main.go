package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tebrizetayi/ledger_service/internal/api"
	"github.com/tebrizetayi/ledger_service/internal/transaction_manager"

	_ "github.com/lib/pq"
)

func main() {

	port := os.Getenv("SERVER_LISTEN_ADDR")
	if port == "" {
		port = ":8080"
	}

	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Use the db object for querying and other operations

	defer db.Close()

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Services
	transaction_manager := transaction_manager.NewTransactionManagerClient(db)
	controller := api.NewController(transaction_manager)

	// Start the HTTP service listening for requests.
	api := http.Server{
		Addr:           port,
		Handler:        api.NewAPI(controller),
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("main : API Listening %s", port)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown
	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("main : Error starting server: %+v", err)

	case sig := <-shutdown:
		log.Printf("main : %v : Start shutdown..", sig)
	}
}
func connectToDatabase() (*sql.DB, error) {
	host := "postgres"
	port := 5432
	user := "postgres"
	password := "postgres"
	dbname := "ledger"

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to the database!")
	return db, nil
}

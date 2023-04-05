package main

import (
	"context"
	"fmt"
	"ledger_service/internal/api"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/allaboutapps/integresql-client-go"
	"github.com/gofrs/uuid"
)

func main() {

	port := os.Getenv("SERVER_LISTEN_ADDR")
	if port == "" {
		port = ":8080"
	}

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Services
	controller := api.NewController()

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

func Z() {
	ctx := context.Background()
	client, err := integresql.DefaultClientFromEnv()
	if err != nil {
		log.Fatalf("Failed to create integresql client: %v", err)
	}

	hash := uuid.Must(uuid.NewV4()).String()

	template, err := client.InitializeTemplate(ctx, hash)
	if err != nil {
		log.Fatalf("Failed to initialize template: %v", err)
	}

	if len(template.Config.Database) == 0 {
		log.Fatalf("Template config database is empty")
	}

	testDB, err := client.GetTestDatabase(ctx, hash)
	if err != nil {
		log.Fatalf("Failed to create test database: %v", err)
	}

	defer func() {
		err := client.DiscardTemplate(ctx, hash)
		if err != nil {
			log.Printf("Failed to drop test database: %v", err)
		}
	}()

	fmt.Println(testDB)
}

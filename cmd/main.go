package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"github.com/tebrizetayi/ledger_service/internal/api"
	"github.com/tebrizetayi/ledger_service/internal/storage"
	"github.com/tebrizetayi/ledger_service/internal/transaction_manager"

	_ "github.com/lib/pq"
)

func main() {
	config := initConfig()
	db, err := connectToDatabase(config.DB)
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
	storageClient := storage.NewStorageClient(db)
	transactionManager := transaction_manager.NewTransactionManagerClient(storageClient)
	controller := api.NewController(transactionManager)

	// Start the HTTP service listening for requests.
	api := http.Server{
		Addr:           config.App.Port,
		Handler:        api.NewAPI(controller),
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("main : API Listening %s", config.App.Port)
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

type Config struct {
	DB  DBConfig
	App AppConfig
}
type AppConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func initConfig() Config {
	viper.AutomaticEnv()

	return Config{
		DB: DBConfig{
			Host:     viper.GetString("POSTGRES_HOST"),
			Port:     viper.GetInt("POSTGRES_PORT"),
			User:     viper.GetString("POSTGRES_USER"),
			Password: viper.GetString("POSTGRES_PASSWORD"),
			DBName:   viper.GetString("POSTGRES_DB"),
			SSLMode:  viper.GetString("POSTGRES_SSLMODE"),
		},
		App: AppConfig{
			Port: viper.GetString("PORT"),
		},
	}
}

func connectToDatabase(dBConfig DBConfig) (*sql.DB, error) {
	connectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dBConfig.Host,
		dBConfig.Port,
		dBConfig.User,
		dBConfig.Password,
		dBConfig.DBName,
		dBConfig.SSLMode,
	)

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

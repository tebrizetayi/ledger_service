package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/allaboutapps/integresql-client-go"
)

type env struct {
	Context context.Context
}

var (
	testEnv env
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	testEnv = env{
		Context: ctx,
	}

	os.Exit(m.Run())
}

func createTestDB(ctx context.Context) (*sql.DB, error) {

	hash := fmt.Sprintf("%d", time.Now().UnixNano())

	client, err := integresql.DefaultClientFromEnv()
	if err != nil {
		return nil, err
	}

	if err := client.ResetAllTracking(ctx); err != nil {
		return nil, err
	}

	if err := client.SetupTemplateWithDBClient(ctx, hash, func(db *sql.DB) error {
		if err := populateTemplateDB(ctx, db); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	testDatabase, err := client.GetTestDatabase(ctx, hash)
	if err != nil {
		log.Fatalf("failed to get test database: %v", err)
	}

	testDb, err := sql.Open("postgres", testDatabase.Config.ConnectionString())
	if err != nil {
		return nil, err
	}
	log.Printf("test database: %s", testDatabase.Config.ConnectionString())

	return testDb, nil
}

func populateTemplateDB(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
	  CREATE TABLE transactions (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		amount DOUBLE PRECISION NOT NULL,
		created_at TIMESTAMP NOT NULL
	);`); err != nil {
		return err
	}

	return nil
}

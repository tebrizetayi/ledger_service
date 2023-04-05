package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/allaboutapps/integresql-client-go"
	_ "github.com/lib/pq"
)

type TestEnv struct {
	Context context.Context
	DB      *sql.DB
}

func PopulateTemplateDB(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS  users (
		id UUID PRIMARY KEY,
		username TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS  transactions (
		id UUID PRIMARY KEY,
		user_id UUID NOT NULL,
		amount DOUBLE PRECISION NOT NULL,
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE);`); err != nil {
		return err
	}

	return nil
}

func CreateTestDB(ctx context.Context) (*sql.DB, error) {

	hash := fmt.Sprintf("%d", time.Now().UnixNano())

	client, err := integresql.DefaultClientFromEnv()
	if err != nil {
		return nil, err
	}

	if err := client.ResetAllTracking(ctx); err != nil {
		return nil, err
	}

	if err := client.SetupTemplateWithDBClient(ctx, hash, func(db *sql.DB) error {
		if err := PopulateTemplateDB(ctx, db); err != nil {
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
	log.Printf("test database: %s", testDatabase.Config.Database)

	return testDb, nil
}

func CreateTestEnv() (TestEnv, error) {

	ctx := context.Background()
	testDb, err := CreateTestDB(ctx)
	if err != nil {
		return TestEnv{}, err
	}

	testEnv := TestEnv{
		Context: ctx,
		DB:      testDb,
	}

	return testEnv, nil
}

func CleanUpTestEnv(testEnv *TestEnv) error {
	if testEnv.DB == nil {
		return nil
	}
	return testEnv.DB.Close()
}

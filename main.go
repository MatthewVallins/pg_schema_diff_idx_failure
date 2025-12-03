package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/stripe/pg-schema-diff/pkg/diff"
	"github.com/stripe/pg-schema-diff/pkg/tempdb"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("ERROR: DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	dbConfig, err := pgx.ParseConfig(dbURL)
	if err != nil {
		fmt.Printf("ERROR: Failed to parse DATABASE_URL: %v\n", err)
		os.Exit(1)
	}

	pool := stdlib.OpenDB(*dbConfig)
	if err := pool.PingContext(ctx); err != nil {
		fmt.Printf("ERROR: Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	pool.SetMaxOpenConns(5)

	tempDbFactory, err := tempdb.NewOnInstanceFactory(
		ctx,
		func(ctx context.Context, dbName string) (*sql.DB, error) {
			copiedConfig := dbConfig.Copy()
			copiedConfig.Database = dbName
			p := stdlib.OpenDB(*copiedConfig)
			if err := p.PingContext(ctx); err != nil {
				p.Close()
				return nil, err
			}
			p.SetMaxOpenConns(5)
			return p, nil
		},
		tempdb.WithRootDatabase(dbConfig.Database),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create temp db factory: %v\n", err)
		os.Exit(1)
	}

	// Test 1: Create expression index WITHOUT WithNoConcurrentIndexOps
	fmt.Println("=== TEST 1: Create expression index without WithNoConcurrentIndexOps ===")
	fmt.Println("From: schema_no_index.sql (table only)")
	fmt.Println("To:   schema_before.sql (table + expression index)")
	fmt.Println("")

	schemaWithIndex, _ := os.ReadFile("schema_before.sql")
	_, err = diff.Generate(
		ctx,
		diff.DBSchemaSource(pool),
		diff.DDLSchemaSource([]string{string(schemaWithIndex)}),
		diff.WithTempDbFactory(tempDbFactory),
	)
	if err != nil {
		fmt.Println("FAILED (expected)")
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("SUCCESS (unexpected)")
	}

	// Test 2: Add column WITH WithNoConcurrentIndexOps (validation still fails)
	fmt.Println("")
	fmt.Println("=== TEST 2: Add column with WithNoConcurrentIndexOps ===")
	fmt.Println("From: schema_before.sql (table + expression index)")
	fmt.Println("To:   schema_after.sql (adds value3 column)")
	fmt.Println("")

	// Reset DB to have the index
	pool.ExecContext(ctx, "DROP TABLE IF EXISTS index_test")
	pool.ExecContext(ctx, string(schemaWithIndex))

	schemaAfter, _ := os.ReadFile("schema_after.sql")
	_, err = diff.Generate(
		ctx,
		diff.DBSchemaSource(pool),
		diff.DDLSchemaSource([]string{string(schemaAfter)}),
		diff.WithTempDbFactory(tempDbFactory),
		diff.WithNoConcurrentIndexOps(),
	)
	if err != nil {
		fmt.Println("FAILED (expected - validation step ignores WithNoConcurrentIndexOps)")
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("SUCCESS (unexpected)")
	}
}

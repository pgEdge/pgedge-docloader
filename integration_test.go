//go:build integration
// +build integration

//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Portions copyright (c) 2025, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/pgedge/pgedge-docloader/internal/database"
	"github.com/pgedge/pgedge-docloader/internal/processor"
	"github.com/pgedge/pgedge-docloader/internal/types"
)

// These integration tests require a running PostgreSQL database
// Run with: go test -tags=integration -v

func getTestConfig() *types.Config {
	return &types.Config{
		DBHost:              getEnv("PGHOST", "localhost"),
		DBPort:              5432,
		DBName:              getEnv("PGDATABASE", "test"),
		DBUser:              getEnv("PGUSER", "postgres"),
		DBPassword:          getEnv("PGPASSWORD", "postgres"),
		DBSSLMode:           "disable",
		DBTable:             "test_documents",
		ColumnDocTitle:      "title",
		ColumnDocContent:    "content",
		ColumnSourceContent: "source",
		ColumnFileName:      "filename",
		ColumnFileModified:  "modified",
		ColumnRowCreated:    "created_at",
		ColumnRowUpdated:    "updated_at",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestIntegrationInsertDocuments(t *testing.T) {
	cfg := getTestConfig()

	// Create database client
	dbClient, err := database.New(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer dbClient.Close()

	// Create test table
	ctx := context.Background()
	if err := createTestTable(ctx, dbClient, cfg); err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	defer dropTestTable(ctx, dbClient, cfg)

	// Process test documents
	docs, stats, err := processor.ProcessFiles("testdata", false)
	if err != nil {
		t.Fatalf("Failed to process files: %v", err)
	}

	// Insert documents
	if err := dbClient.InsertDocuments(ctx, docs, stats); err != nil {
		t.Fatalf("Failed to insert documents: %v", err)
	}

	// Verify inserts
	if stats.FilesInserted == 0 {
		t.Error("Expected documents to be inserted")
	}

	t.Logf("Inserted %d documents", stats.FilesInserted)
}

func TestIntegrationUpdateDocuments(t *testing.T) {
	cfg := getTestConfig()
	cfg.UpdateMode = true

	// Create database client
	dbClient, err := database.New(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer dbClient.Close()

	// Create test table
	ctx := context.Background()
	if err := createTestTable(ctx, dbClient, cfg); err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	defer dropTestTable(ctx, dbClient, cfg)

	// Process and insert documents first time
	docs, stats, err := processor.ProcessFiles("testdata/sample.md", false)
	if err != nil {
		t.Fatalf("Failed to process files: %v", err)
	}

	if err := dbClient.InsertDocuments(ctx, docs, stats); err != nil {
		t.Fatalf("Failed to insert documents: %v", err)
	}

	initialInserts := stats.FilesInserted

	// Process and update documents second time
	stats2 := &types.Stats{}
	if err := dbClient.InsertDocuments(ctx, docs, stats2); err != nil {
		t.Fatalf("Failed to update documents: %v", err)
	}

	if stats2.FilesUpdated == 0 {
		t.Error("Expected documents to be updated")
	}

	if stats2.FilesUpdated != initialInserts {
		t.Errorf("Expected %d updates, got %d", initialInserts, stats2.FilesUpdated)
	}

	t.Logf("Updated %d documents", stats2.FilesUpdated)
}

func createTestTable(ctx context.Context, dbClient *database.Client, cfg *types.Config) error {
	// This is a helper to access the pool - in a real implementation
	// we'd expose a method to execute raw SQL
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	createSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id SERIAL PRIMARY KEY,
            %s TEXT,
            %s TEXT,
            %s BYTEA,
            %s TEXT UNIQUE,
            %s TIMESTAMP,
            %s TIMESTAMP,
            %s TIMESTAMP
        )`,
		cfg.DBTable,
		cfg.ColumnDocTitle,
		cfg.ColumnDocContent,
		cfg.ColumnSourceContent,
		cfg.ColumnFileName,
		cfg.ColumnFileModified,
		cfg.ColumnRowCreated,
		cfg.ColumnRowUpdated,
	)

	_, err = conn.Exec(ctx, createSQL)
	return err
}

func dropTestTable(ctx context.Context, dbClient *database.Client, cfg *types.Config) {
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return
	}
	defer conn.Close(ctx)

	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", cfg.DBTable)
	conn.Exec(ctx, dropSQL)
}

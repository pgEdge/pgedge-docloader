//go:build integration
// +build integration

//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Copyright (c) 2025 - 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"

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

// testConnString builds a connection URL for the test database, escaping
// each component so that unusual characters in the password or database
// name cannot alter the connection parameters.
func testConnString(cfg *types.Config) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:     net.JoinHostPort(cfg.DBHost, strconv.Itoa(cfg.DBPort)),
		Path:     cfg.DBName,
		RawQuery: url.Values{"sslmode": {cfg.DBSSLMode}}.Encode(),
	}
	return u.String()
}

func createTestTable(ctx context.Context, dbClient *database.Client, cfg *types.Config) error {
	conn, err := pgx.Connect(ctx, testConnString(cfg))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, createTestTableSQL(cfg))
	return err
}

// createTestTableSQL builds the CREATE TABLE statement for the test
// table, quoting the configured table and column names as identifiers.
func createTestTableSQL(cfg *types.Config) string {
	return fmt.Sprintf(`
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
		pgx.Identifier{cfg.DBTable}.Sanitize(),
		pgx.Identifier{cfg.ColumnDocTitle}.Sanitize(),
		pgx.Identifier{cfg.ColumnDocContent}.Sanitize(),
		pgx.Identifier{cfg.ColumnSourceContent}.Sanitize(),
		pgx.Identifier{cfg.ColumnFileName}.Sanitize(),
		pgx.Identifier{cfg.ColumnFileModified}.Sanitize(),
		pgx.Identifier{cfg.ColumnRowCreated}.Sanitize(),
		pgx.Identifier{cfg.ColumnRowUpdated}.Sanitize(),
	)
}

// dropTestTableSQL builds the DROP TABLE statement for the test table.
func dropTestTableSQL(cfg *types.Config) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s",
		pgx.Identifier{cfg.DBTable}.Sanitize())
}

func dropTestTable(ctx context.Context, dbClient *database.Client, cfg *types.Config) {
	conn, err := pgx.Connect(ctx, testConnString(cfg))
	if err != nil {
		return
	}
	defer conn.Close(ctx)

	conn.Exec(ctx, dropTestTableSQL(cfg))
}

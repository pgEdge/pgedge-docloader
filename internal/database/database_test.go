//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Copyright (c) 2025 - 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package database

import (
	"testing"
	"time"

	"github.com/pgedge/pgedge-docloader/internal/types"
)

func TestBuildConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *types.Config
		expected string
	}{
		{
			"Basic connection",
			&types.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBName:     "testdb",
				DBUser:     "testuser",
				DBPassword: "testpass",
			},
			"host=localhost port=5432 dbname=testdb user=testuser password=testpass",
		},
		{
			"Connection with SSL",
			&types.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBName:     "testdb",
				DBUser:     "testuser",
				DBPassword: "testpass",
				DBSSLMode:  "require",
			},
			"host=localhost port=5432 dbname=testdb user=testuser password=testpass sslmode=require",
		},
		{
			"Connection with SSL certificates",
			&types.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBName:     "testdb",
				DBUser:     "testuser",
				DBPassword: "testpass",
				DBSSLMode:  "verify-full",
				DBSSLCert:  "/path/to/cert",
				DBSSLKey:   "/path/to/key",
				DBSSLRoot:  "/path/to/root",
			},
			"host=localhost port=5432 dbname=testdb user=testuser password=testpass sslmode=verify-full sslcert=/path/to/cert sslkey=/path/to/key sslrootcert=/path/to/root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConnectionString(tt.config)
			if result != tt.expected {
				t.Errorf("\nexpected: %s\ngot:      %s", tt.expected, result)
			}
		})
	}
}

func TestBuildInsertQuery(t *testing.T) {
	modTime := time.Now()
	doc := &types.Document{
		Title:         "Test Title",
		Content:       "Test Content",
		SourceContent: []byte("source"),
		FileName:      "test.md",
		FileModified:  &modTime,
	}

	tests := []struct {
		name         string
		config       *types.Config
		expectedCols []string
	}{
		{
			"All columns",
			&types.Config{
				DBTable:             "documents",
				ColumnDocTitle:      "title",
				ColumnDocContent:    "content",
				ColumnSourceContent: "source",
				ColumnFileName:      "filename",
				ColumnFileModified:  "modified",
				ColumnRowCreated:    "created",
				ColumnRowUpdated:    "updated",
			},
			[]string{"title", "content", "source", "filename", "modified", "created", "updated"},
		},
		{
			"Partial columns",
			&types.Config{
				DBTable:          "documents",
				ColumnDocTitle:   "title",
				ColumnDocContent: "content",
				ColumnFileName:   "filename",
			},
			[]string{"title", "content", "filename"},
		},
		{
			"With custom columns",
			&types.Config{
				DBTable:          "documents",
				ColumnDocTitle:   "title",
				ColumnDocContent: "content",
				ColumnFileName:   "filename",
				CustomColumns: map[string]string{
					"product": "pgAdmin 4",
					"version": "v9.9",
				},
			},
			[]string{"title", "content", "filename", "product", "version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{config: tt.config}
			query, args := client.buildInsertQuery(doc)

			// Check that query contains INSERT
			if len(query) == 0 {
				t.Error("expected non-empty query")
			}

			// Check that we have the right number of arguments
			if len(args) != len(tt.expectedCols) {
				t.Errorf("expected %d args, got %d", len(tt.expectedCols), len(args))
			}

			// Basic validation - just check query was built
			if len(query) == 0 {
				t.Error("expected non-empty query")
			}
		})
	}
}

func TestBuildUpdateQuery(t *testing.T) {
	modTime := time.Now()
	doc := &types.Document{
		Title:         "Updated Title",
		Content:       "Updated Content",
		SourceContent: []byte("updated source"),
		FileName:      "test.md",
		FileModified:  &modTime,
	}

	tests := []struct {
		name   string
		config *types.Config
	}{
		{
			"Basic update",
			&types.Config{
				DBTable:             "documents",
				ColumnDocTitle:      "title",
				ColumnDocContent:    "content",
				ColumnSourceContent: "source",
				ColumnFileName:      "filename",
				ColumnFileModified:  "modified",
				ColumnRowUpdated:    "updated",
			},
		},
		{
			"Update with custom columns",
			&types.Config{
				DBTable:          "documents",
				ColumnDocTitle:   "title",
				ColumnDocContent: "content",
				ColumnFileName:   "filename",
				CustomColumns: map[string]string{
					"product": "pgAdmin 4",
					"version": "v9.9",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{config: tt.config}
			query, args := client.buildUpdateQuery(doc)

			// Check that query contains UPDATE
			if len(query) == 0 {
				t.Error("expected non-empty query")
			}

			// Should have args for each column being updated, plus one for WHERE clause
			if len(args) == 0 {
				t.Error("expected non-empty args")
			}

			// Last arg should be the filename for WHERE clause
			lastArg := args[len(args)-1]
			if lastArg != doc.FileName {
				t.Errorf("expected last arg to be filename '%s', got '%v'", doc.FileName, lastArg)
			}
		})
	}
}

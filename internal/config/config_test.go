//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Portions copyright (c) 2025, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pgedge/pgedge-docloader/internal/types"
)

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		baseDir  string
		expected string
	}{
		{
			"Empty path",
			"",
			"/base",
			"",
		},
		{
			"Absolute path",
			"/absolute/path",
			"/base",
			"/absolute/path",
		},
		{
			"Relative path",
			"relative/path",
			"/base",
			"/base/relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePath(tt.path, tt.baseDir)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *types.Config
		shouldErr bool
	}{
		{
			"Valid config",
			&types.Config{
				Source:           "/path/to/source",
				DBHost:           "localhost",
				DBName:           "testdb",
				DBUser:           "testuser",
				DBTable:          "testtable",
				ColumnDocContent: "content",
			},
			false,
		},
		{
			"Missing source",
			&types.Config{
				DBHost:           "localhost",
				DBName:           "testdb",
				DBUser:           "testuser",
				DBTable:          "testtable",
				ColumnDocContent: "content",
			},
			true,
		},
		{
			"Missing DB host",
			&types.Config{
				Source:           "/path/to/source",
				DBName:           "testdb",
				DBUser:           "testuser",
				DBTable:          "testtable",
				ColumnDocContent: "content",
			},
			true,
		},
		{
			"Missing DB name",
			&types.Config{
				Source:           "/path/to/source",
				DBHost:           "localhost",
				DBUser:           "testuser",
				DBTable:          "testtable",
				ColumnDocContent: "content",
			},
			true,
		},
		{
			"Missing DB user",
			&types.Config{
				Source:           "/path/to/source",
				DBHost:           "localhost",
				DBName:           "testdb",
				DBTable:          "testtable",
				ColumnDocContent: "content",
			},
			true,
		},
		{
			"Missing DB table",
			&types.Config{
				Source:           "/path/to/source",
				DBHost:           "localhost",
				DBName:           "testdb",
				DBUser:           "testuser",
				ColumnDocContent: "content",
			},
			true,
		},
		{
			"Missing all columns",
			&types.Config{
				Source:  "/path/to/source",
				DBHost:  "localhost",
				DBName:  "testdb",
				DBUser:  "testuser",
				DBTable: "testtable",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if tt.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestReadPgPass(t *testing.T) {
	// Create a temporary .pgpass file
	tmpDir := t.TempDir()
	pgpassFile := filepath.Join(tmpDir, ".pgpass")

	content := `# Test pgpass file
localhost:5432:testdb:testuser:testpass
another:5432:otherdb:otheruser:otherpass
`

	if err := os.WriteFile(pgpassFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create test .pgpass file: %v", err)
	}

	// Temporarily replace home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	password, err := readPgPass()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return the first non-comment password
	if password != "testpass" {
		t.Errorf("expected 'testpass', got '%s'", password)
	}
}

func TestGetPasswordFromEnv(t *testing.T) {
	// Set PGPASSWORD environment variable
	os.Setenv("PGPASSWORD", "envpassword")
	defer os.Unsetenv("PGPASSWORD")

	password, err := getPassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if password != "envpassword" {
		t.Errorf("expected 'envpassword', got '%s'", password)
	}
}

func TestGetPasswordEmpty(t *testing.T) {
	// Ensure no password is set in environment
	os.Unsetenv("PGPASSWORD")

	// Create temp home dir without .pgpass
	originalHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	password, err := getPassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return empty string for passwordless authentication
	if password != "" {
		t.Errorf("expected empty password for passwordless auth, got '%s'", password)
	}
}

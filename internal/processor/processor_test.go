//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Copyright (c) 2025 - 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessFile(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"test.md":   "# Test Title\n\nTest content",
		"test.html": "<html><head><title>HTML Test</title></head><body><p>Content</p></body></html>",
		"test.rst":  "Test Title\n==========\n\nTest content",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	tests := []struct {
		name      string
		filename  string
		stripPath bool
		wantTitle string
		wantErr   bool
	}{
		{
			"Process Markdown file",
			"test.md",
			false,
			"Test Title",
			false,
		},
		{
			"Process HTML file",
			"test.html",
			false,
			"HTML Test",
			false,
		},
		{
			"Process RST file",
			"test.rst",
			false,
			"Test Title",
			false,
		},
		{
			"Strip path",
			"test.md",
			true,
			"Test Title",
			false,
		},
	}

	root, err := os.OpenRoot(tmpDir)
	if err != nil {
		t.Fatalf("failed to open test root: %v", err)
	}
	defer root.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.filename)
			doc, err := processFile(root, tt.filename, filePath, tt.stripPath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if doc.Title != tt.wantTitle {
				t.Errorf("expected title '%s', got '%s'", tt.wantTitle, doc.Title)
			}

			if tt.stripPath {
				if doc.FileName != tt.filename {
					t.Errorf("expected filename '%s', got '%s'", tt.filename, doc.FileName)
				}
			} else {
				if doc.FileName != filePath {
					t.Errorf("expected filename '%s', got '%s'", filePath, doc.FileName)
				}
			}

			if doc.Content == "" {
				t.Error("expected non-empty content")
			}

			if len(doc.SourceContent) == 0 {
				t.Error("expected non-empty source content")
			}

			if doc.FileModified == nil {
				t.Error("expected file modified time")
			}
		})
	}
}

func TestProcessFiles(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"doc1.md":    "# Document 1\n\nContent 1",
		"doc2.md":    "# Document 2\n\nContent 2",
		"doc3.html":  "<html><head><title>Doc 3</title></head><body><p>Content 3</p></body></html>",
		"readme.txt": "This should be skipped",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	t.Run("Process directory", func(t *testing.T) {
		docs, stats, err := ProcessFiles(tmpDir, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should process 3 supported files and skip 1
		if stats.FilesProcessed != 3 {
			t.Errorf("expected 3 files processed, got %d", stats.FilesProcessed)
		}

		if stats.FilesSkipped != 1 {
			t.Errorf("expected 1 file skipped, got %d", stats.FilesSkipped)
		}

		if len(docs) != 3 {
			t.Errorf("expected 3 documents, got %d", len(docs))
		}
	})

	t.Run("Process glob pattern", func(t *testing.T) {
		pattern := filepath.Join(tmpDir, "*.md")
		docs, stats, err := ProcessFiles(pattern, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should process only .md files
		if stats.FilesProcessed != 2 {
			t.Errorf("expected 2 files processed, got %d", stats.FilesProcessed)
		}

		if len(docs) != 2 {
			t.Errorf("expected 2 documents, got %d", len(docs))
		}
	})

	t.Run("Process single file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "doc1.md")
		docs, stats, err := ProcessFiles(filePath, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stats.FilesProcessed != 1 {
			t.Errorf("expected 1 file processed, got %d", stats.FilesProcessed)
		}

		if len(docs) != 1 {
			t.Errorf("expected 1 document, got %d", len(docs))
		}
	})

	t.Run("Unsupported single file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "readme.txt")
		_, _, err := ProcessFiles(filePath, false)
		if err == nil {
			t.Error("expected error for unsupported file, got nil")
		}
	})

	t.Run("Recursive glob pattern", func(t *testing.T) {
		// Create nested directory structure
		nestedDir := t.TempDir()
		subdirs := []string{"", "subdir1", "subdir2", "subdir2/nested"}

		expectedFiles := 0
		for _, subdir := range subdirs {
			dir := filepath.Join(nestedDir, subdir)
			if subdir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("failed to create directory %s: %v", dir, err)
				}
			} else {
				dir = nestedDir
			}

			// Create a markdown file in each directory
			mdFile := filepath.Join(dir, "doc.md")
			if err := os.WriteFile(mdFile, []byte("# Test\n\nContent"), 0644); err != nil {
				t.Fatalf("failed to create file %s: %v", mdFile, err)
			}
			expectedFiles++

			// Also create a non-markdown file that should be skipped
			txtFile := filepath.Join(dir, "readme.txt")
			if err := os.WriteFile(txtFile, []byte("Ignore this"), 0644); err != nil {
				t.Fatalf("failed to create file %s: %v", txtFile, err)
			}
		}

		// Test recursive glob pattern
		pattern := filepath.Join(nestedDir, "**/*.md")
		docs, stats, err := ProcessFiles(pattern, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stats.FilesProcessed != expectedFiles {
			t.Errorf("expected %d files processed, got %d", expectedFiles, stats.FilesProcessed)
		}

		if len(docs) != expectedFiles {
			t.Errorf("expected %d documents, got %d", expectedFiles, len(docs))
		}
	})
}

func TestSourceRoot(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{"Plain directory", "docs", "docs"},
		{"Trailing wildcard", filepath.Join("docs", "*.md"), "docs"},
		{"Recursive wildcard", filepath.Join("docs", "**", "*.md"), "docs"},
		{"Wildcard in directory", filepath.Join("docs", "*", "index.md"), "docs"},
		{"Bare wildcard", "*.md", "."},
		{"Character class", filepath.Join("docs", "[ab]*.md"), "docs"},
		{"Dot-slash prefix", filepath.Join(".", "*.md"), "."},
		{"Filesystem root", string(filepath.Separator), string(filepath.Separator)},
		{"Wildcard at filesystem root", string(filepath.Separator) + "*.md", string(filepath.Separator)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourceRoot(tt.source); got != tt.expected {
				t.Errorf("expected root %q, got %q", tt.expected, got)
			}
		})
	}
}

// A glob whose directory does not exist matches nothing. It must not fail the
// run, because a caller loading several source paths would lose all of them.
func TestProcessFilesGlobWithMissingDirectory(t *testing.T) {
	pattern := filepath.Join(t.TempDir(), "absent", "*.md")

	docs, stats, err := ProcessFiles(pattern, false)
	if err != nil {
		t.Fatalf("expected a missing glob directory to match nothing, got: %v", err)
	}

	if len(docs) != 0 || stats.FilesProcessed != 0 {
		t.Errorf("expected no documents, got %d docs and %d processed", len(docs), stats.FilesProcessed)
	}
}

// A symlink pointing outside the source tree must be reported as an error and
// its target must never appear in the loaded documents.
func TestProcessFilesRejectsSymlinkEscape(t *testing.T) {
	secretDir := t.TempDir()
	secret := filepath.Join(secretDir, "secret.md")
	if err := os.WriteFile(secret, []byte("# Secret\n\ntop secret"), 0600); err != nil {
		t.Fatalf("failed to create secret file: %v", err)
	}

	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "ok.md"), []byte("# Ok\n\nfine"), 0600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}
	if err := os.Symlink(secret, filepath.Join(sourceDir, "escape.md")); err != nil {
		t.Skipf("symlinks unsupported on this platform: %v", err)
	}

	docs, stats, err := ProcessFiles(sourceDir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.FilesProcessed != 1 {
		t.Errorf("expected only the in-tree file to be processed, got %d", stats.FilesProcessed)
	}

	if !stats.HasErrors() {
		t.Error("expected the escaping symlink to be recorded as an error")
	}

	for _, doc := range docs {
		if strings.Contains(string(doc.SourceContent), "top secret") {
			t.Errorf("content from outside the source tree was loaded: %s", doc.FileName)
		}
	}
}

// A relative symlink that stays inside the source tree is a legitimate
// documentation layout and must still be loaded.
func TestProcessFilesAllowsSymlinkWithinTree(t *testing.T) {
	sourceDir := t.TempDir()

	shared := filepath.Join(sourceDir, "shared")
	if err := os.MkdirAll(shared, 0750); err != nil {
		t.Fatalf("failed to create shared directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(shared, "doc.md"), []byte("# Shared\n\nreusable"), 0600); err != nil {
		t.Fatalf("failed to create shared file: %v", err)
	}

	link := filepath.Join(sourceDir, "linked.md")
	if err := os.Symlink(filepath.Join("shared", "doc.md"), link); err != nil {
		t.Skipf("symlinks unsupported on this platform: %v", err)
	}

	_, stats, err := ProcessFiles(sourceDir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.HasErrors() {
		t.Errorf("in-tree symlink should load cleanly, got errors: %v", stats.Errors)
	}

	// The shared file plus the symlink pointing at it
	if stats.FilesProcessed != 2 {
		t.Errorf("expected 2 files processed, got %d", stats.FilesProcessed)
	}
}

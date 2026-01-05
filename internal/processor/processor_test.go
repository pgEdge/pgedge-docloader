//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Portions copyright (c) 2025 - 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package processor

import (
	"os"
	"path/filepath"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.filename)
			doc, err := processFile(filePath, tt.stripPath)

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

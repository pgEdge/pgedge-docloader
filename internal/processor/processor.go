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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pgedge/pgedge-docloader/internal/converter"
	"github.com/pgedge/pgedge-docloader/internal/types"
)

// ProcessFiles processes files from the source path
func ProcessFiles(source string, stripPath bool) ([]*types.Document, *types.Stats, error) {
	stats := &types.Stats{}
	var documents []*types.Document

	// Check if source is a single file, directory, or glob pattern
	fileInfo, err := os.Stat(source)
	if err == nil && !fileInfo.IsDir() {
		// Single file
		doc, err := processFile(source, stripPath)
		if err != nil {
			if err == converter.ErrUnsupportedFormat {
				return nil, nil, fmt.Errorf("unsupported file type: %s", source)
			}
			return nil, nil, err
		}
		documents = append(documents, doc)
		stats.FilesProcessed++
	} else {
		// Directory or glob pattern
		var files []string

		// Check if it's a glob pattern
		if strings.ContainsAny(source, "*?[]") {
			// Check for ** recursive glob pattern
			if strings.Contains(source, "**") {
				matches, err := recursiveGlob(source)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to process glob pattern: %w", err)
				}
				files = matches
			} else {
				// Use standard glob for non-recursive patterns
				matches, err := filepath.Glob(source)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid glob pattern: %w", err)
				}
				files = matches
			}
		} else {
			// Directory - walk it recursively
			err := filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read directory: %w", err)
			}
		}

		// Process each file
		for _, file := range files {
			if !converter.IsSupported(file) {
				fmt.Printf("Skipping unsupported file: %s\n", file)
				stats.FilesSkipped++
				continue
			}

			doc, err := processFile(file, stripPath)
			if err != nil {
				fmt.Printf("Error processing file %s: %v\n", file, err)
				stats.AddError(fmt.Errorf("file %s: %w", file, err))
				stats.FilesSkipped++
				continue
			}

			documents = append(documents, doc)
			stats.FilesProcessed++
		}
	}

	return documents, stats, nil
}

// recursiveGlob implements recursive glob matching with ** support
func recursiveGlob(pattern string) ([]string, error) {
	// Split pattern at /** to get base dir and file pattern
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid recursive glob pattern: %s", pattern)
	}

	baseDir := strings.TrimSuffix(parts[0], "/")
	filePattern := strings.TrimPrefix(parts[1], "/")

	// If baseDir is empty, use current directory
	if baseDir == "" {
		baseDir = "."
	}

	var matches []string
	err := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Match the file against the pattern
		matched, err := filepath.Match(filePattern, filepath.Base(path))
		if err != nil {
			return err
		}

		if matched {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return matches, nil
}

// processFile processes a single file
func processFile(filePath string, stripPath bool) (*types.Document, error) {
	// Read file content
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	sourceContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect document type
	docType := converter.DetectDocumentType(filePath)
	if docType == types.TypeUnknown {
		return nil, converter.ErrUnsupportedFormat
	}

	// Convert to markdown
	markdown, title, err := converter.Convert(sourceContent, docType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert document: %w", err)
	}

	// Get file metadata
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Extract timestamps
	modTime := fileInfo.ModTime()

	// Get creation time (platform-specific)
	var createTime *time.Time
	if ct := getCreationTime(fileInfo); ct != nil {
		createTime = ct
	}

	// Determine filename (with or without path)
	fileName := filePath
	if stripPath {
		fileName = filepath.Base(filePath)
	}

	doc := &types.Document{
		Title:         title,
		Content:       markdown,
		SourceContent: sourceContent,
		FileName:      fileName,
		FileCreated:   createTime,
		FileModified:  &modTime,
		DocumentType:  docType,
	}

	return doc, nil
}

// getCreationTime attempts to extract file creation time
func getCreationTime(fileInfo os.FileInfo) *time.Time {
	// This is platform-specific and might not work on all systems
	// On some systems, we can extract it from Sys()
	// For simplicity, we'll return nil if not available
	// A full implementation would use platform-specific syscalls

	// For now, return nil - creation time extraction is optional
	return nil
}

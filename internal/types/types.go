//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Copyright (c) 2025 - 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package types

import "time"

// DocumentType represents the type of source document
type DocumentType int

const (
	TypeUnknown DocumentType = iota
	TypeHTML
	TypeMarkdown
	TypeReStructuredText
	TypeSGML
)

// String returns the string representation of the DocumentType
func (dt DocumentType) String() string {
	switch dt {
	case TypeHTML:
		return "HTML"
	case TypeMarkdown:
		return "Markdown"
	case TypeReStructuredText:
		return "reStructuredText"
	case TypeSGML:
		return "SGML/DocBook"
	default:
		return "Unknown"
	}
}

// Document represents a processed document with all extracted metadata
type Document struct {
	Title         string
	Content       string
	SourceContent []byte
	FileName      string
	FileCreated   *time.Time
	FileModified  *time.Time
	DocumentType  DocumentType
}

// Config represents the application configuration
type Config struct {
	// Source configuration - Local (mutually exclusive with Git source)
	Source    []string // Source paths/patterns (supports multiple via repeated flag or YAML list)
	StripPath bool

	// Source configuration - Git (mutually exclusive with local source)
	GitURL       string   // Git repository URL
	GitBranch    string   // Branch to checkout (mutually exclusive with GitTag)
	GitTag       string   // Tag to checkout (mutually exclusive with GitBranch)
	GitDocPath   []string // Paths within repo to process (supports multiple patterns)
	GitCloneDir  string   // Directory to store cloned repos (default: temp)
	GitKeepClone bool     // Keep cloned repo after processing
	GitSkipFetch bool     // Skip fetch if repo already exists

	// Database configuration
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
	DBTable    string

	// SSL/TLS configuration
	DBSSLCert string
	DBSSLKey  string
	DBSSLRoot string

	// Column mapping
	ColumnDocTitle      string
	ColumnDocContent    string
	ColumnSourceContent string
	ColumnFileName      string
	ColumnFileCreated   string
	ColumnFileModified  string
	ColumnRowCreated    string
	ColumnRowUpdated    string

	// Custom metadata columns (column name -> value)
	CustomColumns map[string]string

	// Operation mode
	UpdateMode bool

	// Configuration file path
	ConfigFile string
}

// Stats tracks processing statistics
type Stats struct {
	FilesProcessed int
	FilesSkipped   int
	FilesInserted  int
	FilesUpdated   int
	Errors         []error
}

// AddError adds an error to the stats
func (s *Stats) AddError(err error) {
	s.Errors = append(s.Errors, err)
}

// HasErrors returns true if there are any errors
func (s *Stats) HasErrors() bool {
	return len(s.Errors) > 0
}

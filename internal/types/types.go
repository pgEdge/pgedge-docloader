//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Portions copyright (c) 2025, pgEdge, Inc.
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
	// Source configuration
	Source    string
	StripPath bool

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

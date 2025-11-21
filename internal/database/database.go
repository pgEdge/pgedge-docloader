//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Portions copyright (c) 2025, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pgedge/pgedge-docloader/internal/types"
)

// Client represents a database client
type Client struct {
	pool   *pgxpool.Pool
	config *types.Config
}

// New creates a new database client
func New(cfg *types.Config) (*Client, error) {
	// Build connection string
	connStr := buildConnectionString(cfg)

	// Create connection pool
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{
		pool:   pool,
		config: cfg,
	}, nil
}

// Close closes the database connection
func (c *Client) Close() {
	c.pool.Close()
}

// InsertDocuments inserts or updates documents in the database
func (c *Client) InsertDocuments(ctx context.Context, documents []*types.Document, stats *types.Stats) error {
	// Begin transaction
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx) //nolint:errcheck // Rollback on defer is safe to ignore
	}()

	for _, doc := range documents {
		if c.config.UpdateMode {
			// Try update first, then insert if not found
			updated, err := c.updateDocument(ctx, tx, doc)
			if err != nil {
				return err
			}

			if updated {
				stats.FilesUpdated++
			} else {
				if err := c.insertDocument(ctx, tx, doc); err != nil {
					return err
				}
				stats.FilesInserted++
			}
		} else {
			// Insert only
			if err := c.insertDocument(ctx, tx, doc); err != nil {
				return err
			}
			stats.FilesInserted++
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// insertDocument inserts a document into the database
func (c *Client) insertDocument(ctx context.Context, tx pgx.Tx, doc *types.Document) error {
	query, args := c.buildInsertQuery(doc)

	_, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

// updateDocument updates a document in the database if it exists
func (c *Client) updateDocument(ctx context.Context, tx pgx.Tx, doc *types.Document) (bool, error) {
	// First check if document exists
	if c.config.ColumnFileName == "" {
		return false, nil
	}

	checkQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1",
		pgx.Identifier{c.config.DBTable}.Sanitize(),
		pgx.Identifier{c.config.ColumnFileName}.Sanitize())

	var count int
	err := tx.QueryRow(ctx, checkQuery, doc.FileName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check document existence: %w", err)
	}

	if count == 0 {
		return false, nil
	}

	// Build update query
	query, args := c.buildUpdateQuery(doc)

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to update document: %w", err)
	}

	return true, nil
}

// buildInsertQuery builds an INSERT query
func (c *Client) buildInsertQuery(doc *types.Document) (string, []interface{}) {
	var columns []string
	var placeholders []string
	var args []interface{}
	argIndex := 1

	// Build column list and values based on configuration
	if c.config.ColumnDocTitle != "" {
		columns = append(columns, pgx.Identifier{c.config.ColumnDocTitle}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, doc.Title)
		argIndex++
	}

	if c.config.ColumnDocContent != "" {
		columns = append(columns, pgx.Identifier{c.config.ColumnDocContent}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, doc.Content)
		argIndex++
	}

	if c.config.ColumnSourceContent != "" {
		columns = append(columns, pgx.Identifier{c.config.ColumnSourceContent}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, doc.SourceContent)
		argIndex++
	}

	if c.config.ColumnFileName != "" {
		columns = append(columns, pgx.Identifier{c.config.ColumnFileName}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, doc.FileName)
		argIndex++
	}

	if c.config.ColumnFileCreated != "" && doc.FileCreated != nil {
		columns = append(columns, pgx.Identifier{c.config.ColumnFileCreated}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, doc.FileCreated)
		argIndex++
	}

	if c.config.ColumnFileModified != "" && doc.FileModified != nil {
		columns = append(columns, pgx.Identifier{c.config.ColumnFileModified}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, doc.FileModified)
		argIndex++
	}

	if c.config.ColumnRowCreated != "" {
		columns = append(columns, pgx.Identifier{c.config.ColumnRowCreated}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, time.Now())
		argIndex++
	}

	if c.config.ColumnRowUpdated != "" {
		columns = append(columns, pgx.Identifier{c.config.ColumnRowUpdated}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, time.Now())
		argIndex++
	}

	// Add custom metadata columns
	for colName, colValue := range c.config.CustomColumns {
		columns = append(columns, pgx.Identifier{colName}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		args = append(args, colValue)
		argIndex++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		pgx.Identifier{c.config.DBTable}.Sanitize(),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return query, args
}

// buildUpdateQuery builds an UPDATE query
func (c *Client) buildUpdateQuery(doc *types.Document) (string, []interface{}) {
	var setClauses []string
	var args []interface{}
	argIndex := 1

	// Build SET clauses based on configuration
	if c.config.ColumnDocTitle != "" {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d",
			pgx.Identifier{c.config.ColumnDocTitle}.Sanitize(), argIndex))
		args = append(args, doc.Title)
		argIndex++
	}

	if c.config.ColumnDocContent != "" {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d",
			pgx.Identifier{c.config.ColumnDocContent}.Sanitize(), argIndex))
		args = append(args, doc.Content)
		argIndex++
	}

	if c.config.ColumnSourceContent != "" {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d",
			pgx.Identifier{c.config.ColumnSourceContent}.Sanitize(), argIndex))
		args = append(args, doc.SourceContent)
		argIndex++
	}

	if c.config.ColumnFileModified != "" && doc.FileModified != nil {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d",
			pgx.Identifier{c.config.ColumnFileModified}.Sanitize(), argIndex))
		args = append(args, doc.FileModified)
		argIndex++
	}

	if c.config.ColumnRowUpdated != "" {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d",
			pgx.Identifier{c.config.ColumnRowUpdated}.Sanitize(), argIndex))
		args = append(args, time.Now())
		argIndex++
	}

	// Add custom metadata columns
	for colName, colValue := range c.config.CustomColumns {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d",
			pgx.Identifier{colName}.Sanitize(), argIndex))
		args = append(args, colValue)
		argIndex++
	}

	// Add WHERE clause
	args = append(args, doc.FileName)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d",
		pgx.Identifier{c.config.DBTable}.Sanitize(),
		strings.Join(setClauses, ", "),
		pgx.Identifier{c.config.ColumnFileName}.Sanitize(),
		argIndex)

	return query, args
}

// buildConnectionString builds a PostgreSQL connection string
func buildConnectionString(cfg *types.Config) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("host=%s", cfg.DBHost))
	parts = append(parts, fmt.Sprintf("port=%d", cfg.DBPort))
	parts = append(parts, fmt.Sprintf("dbname=%s", cfg.DBName))
	parts = append(parts, fmt.Sprintf("user=%s", cfg.DBUser))

	if cfg.DBPassword != "" {
		parts = append(parts, fmt.Sprintf("password=%s", cfg.DBPassword))
	}

	if cfg.DBSSLMode != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", cfg.DBSSLMode))
	}

	if cfg.DBSSLCert != "" {
		parts = append(parts, fmt.Sprintf("sslcert=%s", cfg.DBSSLCert))
	}

	if cfg.DBSSLKey != "" {
		parts = append(parts, fmt.Sprintf("sslkey=%s", cfg.DBSSLKey))
	}

	if cfg.DBSSLRoot != "" {
		parts = append(parts, fmt.Sprintf("sslrootcert=%s", cfg.DBSSLRoot))
	}

	return strings.Join(parts, " ")
}

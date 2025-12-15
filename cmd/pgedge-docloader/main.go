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

	"github.com/spf13/cobra"

	"github.com/pgedge/pgedge-docloader/internal/config"
	"github.com/pgedge/pgedge-docloader/internal/converter"
	"github.com/pgedge/pgedge-docloader/internal/database"
	"github.com/pgedge/pgedge-docloader/internal/processor"
	"github.com/pgedge/pgedge-docloader/internal/types"
)

var (
	version = "1.0.0-beta1"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "pgedge-docloader",
	Short: "pgEdge Document Loader - Load documents into PostgreSQL",
	Long: `pgEdge Document Loader is a tool to load documents from various formats
(HTML, Markdown, reStructuredText, SGML/DocBook) into a PostgreSQL database table.

The tool converts documents to Markdown format and extracts metadata before
storing them in the specified database table.`,
	RunE: run,
}

func init() {
	// Configuration file
	rootCmd.Flags().StringP("config", "c", "", "Path to configuration file")

	// Source configuration
	rootCmd.Flags().StringP("source", "s", "", "Source file, directory, or glob pattern")
	rootCmd.Flags().Bool("strip-path", false, "Strip path from filename, keeping only the base name")

	// Database connection
	rootCmd.Flags().String("db-host", "localhost", "Database host")
	rootCmd.Flags().Int("db-port", 5432, "Database port")
	rootCmd.Flags().String("db-name", "", "Database name")
	rootCmd.Flags().String("db-user", "", "Database user")
	rootCmd.Flags().String("db-sslmode", "prefer", "SSL mode (disable, allow, prefer, require, verify-ca, verify-full)")
	rootCmd.Flags().String("db-table", "", "Database table name")

	// SSL/TLS configuration
	rootCmd.Flags().String("db-sslcert", "", "Path to client SSL certificate")
	rootCmd.Flags().String("db-sslkey", "", "Path to client SSL key")
	rootCmd.Flags().String("db-sslrootcert", "", "Path to SSL root certificate")

	// Column mappings
	rootCmd.Flags().String("col-doc-title", "", "Column name for document title")
	rootCmd.Flags().String("col-doc-content", "", "Column name for document content (markdown)")
	rootCmd.Flags().String("col-source-content", "", "Column name for source content (bytea)")
	rootCmd.Flags().String("col-file-name", "", "Column name for file name")
	rootCmd.Flags().String("col-file-created", "", "Column name for file creation timestamp")
	rootCmd.Flags().String("col-file-modified", "", "Column name for file modification timestamp")
	rootCmd.Flags().String("col-row-created", "", "Column name for row creation timestamp")
	rootCmd.Flags().String("col-row-updated", "", "Column name for row update timestamp")

	// Custom metadata columns
	rootCmd.Flags().StringSlice("set-column", []string{}, "Set custom column value (format: column=value, can be specified multiple times)")

	// Operation mode
	rootCmd.Flags().BoolP("update", "u", false, "Update existing rows (matched by filename) or insert new ones")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("pgedge-docloader %s (commit: %s, built: %s)\n", version, commit, date)
		},
	})

	// List supported formats command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "formats",
		Short: "List supported document formats",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Supported document formats:")
			for _, ext := range converter.GetSupportedExtensions() {
				fmt.Printf("  %s\n", ext)
			}
		},
	})
}

func run(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Process files
	fmt.Printf("Processing files from: %s\n", cfg.Source)
	documents, stats, err := processor.ProcessFiles(cfg.Source, cfg.StripPath)
	if err != nil {
		return fmt.Errorf("failed to process files: %w", err)
	}

	if len(documents) == 0 {
		fmt.Println("No documents to process.")
		return nil
	}

	fmt.Printf("Processed %d file(s), skipped %d file(s)\n",
		stats.FilesProcessed, stats.FilesSkipped)

	// Connect to database
	fmt.Printf("Connecting to database %s@%s:%d/%s\n",
		cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	dbClient, err := database.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbClient.Close()

	// Insert documents
	ctx := context.Background()
	if err := dbClient.InsertDocuments(ctx, documents, stats); err != nil {
		return fmt.Errorf("failed to insert documents: %w", err)
	}

	// Print summary
	printSummary(stats)

	return nil
}

func printSummary(stats *types.Stats) {
	fmt.Println("\n=== Processing Summary ===")
	fmt.Printf("Files processed: %d\n", stats.FilesProcessed)
	fmt.Printf("Files skipped:   %d\n", stats.FilesSkipped)
	fmt.Printf("Rows inserted:   %d\n", stats.FilesInserted)
	fmt.Printf("Rows updated:    %d\n", stats.FilesUpdated)

	if stats.HasErrors() {
		fmt.Printf("\nErrors encountered: %d\n", len(stats.Errors))
		for i, err := range stats.Errors {
			fmt.Printf("  %d. %v\n", i+1, err)
		}
	}

	fmt.Println("=========================")
}

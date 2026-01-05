//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Portions copyright (c) 2025 - 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pgedge/pgedge-docloader/internal/types"
)

// Load loads configuration from file and CLI flags
func Load(cmd *cobra.Command) (*types.Config, error) {
	cfg := &types.Config{}

	// Get config file path if specified
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, fmt.Errorf("failed to get config flag: %w", err)
	}
	if configFile != "" {
		cfg.ConfigFile = configFile

		// Set the config file path
		viper.SetConfigFile(configFile)

		// Read the config file
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Bind CLI flags to viper
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Load configuration values (CLI flags override config file)
	cfg.Source = viper.GetString("source")
	cfg.StripPath = viper.GetBool("strip-path")

	cfg.DBHost = viper.GetString("db-host")
	cfg.DBPort = viper.GetInt("db-port")
	cfg.DBName = viper.GetString("db-name")
	cfg.DBUser = viper.GetString("db-user")
	cfg.DBSSLMode = viper.GetString("db-sslmode")
	cfg.DBTable = viper.GetString("db-table")

	cfg.DBSSLCert = viper.GetString("db-sslcert")
	cfg.DBSSLKey = viper.GetString("db-sslkey")
	cfg.DBSSLRoot = viper.GetString("db-sslrootcert")

	cfg.ColumnDocTitle = viper.GetString("col-doc-title")
	cfg.ColumnDocContent = viper.GetString("col-doc-content")
	cfg.ColumnSourceContent = viper.GetString("col-source-content")
	cfg.ColumnFileName = viper.GetString("col-file-name")
	cfg.ColumnFileCreated = viper.GetString("col-file-created")
	cfg.ColumnFileModified = viper.GetString("col-file-modified")
	cfg.ColumnRowCreated = viper.GetString("col-row-created")
	cfg.ColumnRowUpdated = viper.GetString("col-row-updated")

	// Parse custom columns from --set-column flags and config file
	cfg.CustomColumns = make(map[string]string)

	// First, try to load from config file as a map
	if viper.IsSet("custom-columns") {
		customCols := viper.GetStringMapString("custom-columns")
		for k, v := range customCols {
			cfg.CustomColumns[k] = v
		}
	}

	// Then, load from CLI flags (which override config file)
	setColumnFlags := viper.GetStringSlice("set-column")
	for _, colValue := range setColumnFlags {
		parts := strings.SplitN(colValue, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid set-column format '%s': expected column=value", colValue)
		}
		colName := strings.TrimSpace(parts[0])
		colVal := strings.TrimSpace(parts[1])
		if colName == "" {
			return nil, fmt.Errorf("invalid set-column format '%s': column name cannot be empty", colValue)
		}
		cfg.CustomColumns[colName] = colVal
	}

	cfg.UpdateMode = viper.GetBool("update")

	// Resolve relative paths relative to config file directory
	if cfg.ConfigFile != "" {
		configDir := filepath.Dir(cfg.ConfigFile)
		cfg.Source = resolvePath(cfg.Source, configDir)
		cfg.DBSSLCert = resolvePath(cfg.DBSSLCert, configDir)
		cfg.DBSSLKey = resolvePath(cfg.DBSSLKey, configDir)
		cfg.DBSSLRoot = resolvePath(cfg.DBSSLRoot, configDir)
	}

	// Get password (in order of priority)
	password, err := getPassword()
	if err != nil {
		return nil, err
	}
	cfg.DBPassword = password

	// Validate configuration
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// resolvePath resolves a path relative to a base directory if not absolute
func resolvePath(path, baseDir string) string {
	if path == "" {
		return path
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}

// getPassword gets the database password from various sources
func getPassword() (string, error) {
	// 1. Check PGPASSWORD environment variable
	if password := os.Getenv("PGPASSWORD"); password != "" {
		return password, nil
	}

	// 2. Check .pgpass file
	password, err := readPgPass()
	if err == nil && password != "" {
		return password, nil
	}

	// 3. Return empty string to allow passwordless authentication
	// PostgreSQL supports various authentication methods that don't require passwords:
	// - trust authentication
	// - peer authentication
	// - certificate authentication
	// If a password is actually required and not provided, PostgreSQL will return
	// an authentication error with a clear message.
	return "", nil
}

// readPgPass reads password from .pgpass file
func readPgPass() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	pgpassFile := filepath.Join(homeDir, ".pgpass")
	file, err := os.Open(pgpassFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read .pgpass file
	// Format: hostname:port:database:username:password
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 5 {
			continue
		}

		// For now, return the first matching entry
		// A full implementation would match host:port:database:username
		return parts[4], nil
	}

	return "", scanner.Err()
}

// validate validates the configuration
func validate(cfg *types.Config) error {
	if cfg.Source == "" {
		return fmt.Errorf("source path is required")
	}

	if cfg.DBHost == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if cfg.DBUser == "" {
		return fmt.Errorf("database user is required")
	}

	if cfg.DBTable == "" {
		return fmt.Errorf("database table is required")
	}

	// At least one column must be specified
	if cfg.ColumnDocTitle == "" &&
		cfg.ColumnDocContent == "" &&
		cfg.ColumnSourceContent == "" &&
		cfg.ColumnFileName == "" &&
		cfg.ColumnFileCreated == "" &&
		cfg.ColumnFileModified == "" &&
		cfg.ColumnRowCreated == "" &&
		cfg.ColumnRowUpdated == "" {
		return fmt.Errorf("at least one column mapping must be specified")
	}

	return nil
}

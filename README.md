# pgEdge Document Loader

[![CI](https://github.com/pgEdge/pgedge-docloader/actions/workflows/ci.yml/badge.svg)](https://github.com/pgEdge/pgedge-docloader/actions/workflows/ci.yml)

A command-line tool for loading documents from various formats into PostgreSQL
databases.

## Overview

The pgEdge Document Loader automatically converts documents (HTML, Markdown,
and reStructuredText) to Markdown format and loads them into a PostgreSQL
database with extracted metadata.

## Features

- **Multiple Format Support**: HTML, Markdown, and reStructuredText
- **Automatic Conversion**: All formats converted to Markdown
- **Metadata Extraction**: Titles, filenames, timestamps
- **Flexible Input**: Single file, directory, or glob patterns (including `**` recursive matching)
- **Database Flexibility**: Configurable column mappings
- **Custom Metadata Columns**: Add fixed values to custom columns for every row
- **Update Mode**: Update existing rows or insert new ones
- **Transactional**: All-or-nothing processing with automatic rollback
- **Secure**: Password from environment, .pgpass, or interactive prompt
- **Configuration Files**: Reusable YAML configuration

## Quick Start

### Installation

```bash
git clone https://github.com/pgedge/pgedge-docloader.git
cd pgedge-docloader
make build
make install
```

### Basic Usage

```bash
# Load Markdown files into PostgreSQL
pgedge-docloader \
  --source ./docs \
  --db-host localhost \
  --db-name mydb \
  --db-user myuser \
  --db-table documents \
  --col-doc-content content \
  --col-file-name filename
```

### Using Configuration File

```bash
# Create config.yml
cat > config.yml <<EOF
source: "./docs"
db-host: localhost
db-name: mydb
db-user: myuser
db-table: documents
col-doc-content: content
col-file-name: filename
update: true
EOF

# Run with config
export PGPASSWORD=mypassword
pgedge-docloader --config config.yml
```

## Supported Formats

- **HTML** (`.html`, `.htm`) - Extracts title from `<title>` tag
- **Markdown** (`.md`) - Extracts title from first `#` heading
- **reStructuredText** (`.rst`) - Converts to Markdown

## Documentation

Full documentation is available at:
[https://pgedge.github.io/pgedge-docloader](https://pgedge.github.io/pgedge-docloader)

- [Installation Guide](docs/installation.md)
- [Configuration](docs/configuration.md)
- [Usage Examples](docs/usage.md)
- [Database Setup](docs/database-setup.md)
- [Supported Formats](docs/supported-formats.md)
- [Troubleshooting](docs/troubleshooting.md)

## Requirements

- Go 1.21 or later
- PostgreSQL 12 or later

## Database Setup

Create a table with appropriate columns:

```sql
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    title TEXT,
    content TEXT NOT NULL,
    source BYTEA,
    filename TEXT UNIQUE,
    modified TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Map columns using CLI flags or configuration file.

## Command-Line Options

```
Flags:
  -c, --config string              Path to configuration file
  -s, --source string              Source file, directory, or glob pattern
      --strip-path                 Strip path from filename
      --db-host string             Database host (default "localhost")
      --db-port int                Database port (default 5432)
      --db-name string             Database name
      --db-user string             Database user
      --db-table string            Database table name
      --col-doc-title string       Column for document title
      --col-doc-content string     Column for document content
      --col-source-content string  Column for source content
      --col-file-name string       Column for file name
      --col-file-modified string   Column for file modified timestamp
      --col-row-created string     Column for row created timestamp
      --col-row-updated string     Column for row updated timestamp
  -u, --update                     Update existing rows or insert new ones
```

## Examples

### Load a directory

```bash
pgedge-docloader --source ./documentation --config config.yml
```

### Load with glob pattern

```bash
pgedge-docloader --source "./docs/**/*.md" --config config.yml
```

### Update mode

```bash
pgedge-docloader --source ./docs --config config.yml --update
```

## Development

### Building

```bash
make build
```

The binary will be created at `bin/pgedge-docloader`. To test locally without
installing:

```bash
./bin/pgedge-docloader --help
```

### Running Tests

```bash
make test
```

### Linting

```bash
make lint
```

## License

This project is licensed under the PostgreSQL License. See
[LICENCE.md](LICENCE.md) for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## Support

- Documentation: [https://pgedge.github.io/pgedge-docloader](https://pgedge.github.io/pgedge-docloader)
- Issues: [GitHub Issues](https://github.com/pgedge/pgedge-docloader/issues)

## Project Status

This project is under active development. See the documentation for the latest
features and updates.

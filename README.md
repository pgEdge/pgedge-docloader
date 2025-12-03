# pgEdge Document Loader

[![CI](https://github.com/pgEdge/pgedge-docloader/actions/workflows/ci.yml/badge.svg)](https://github.com/pgEdge/pgedge-docloader/actions/workflows/ci.yml)

## Table of Contents
- [Installation Guide](docs/installation.md)
- [Configuration](docs/configuration.md)
- [Usage Examples](docs/usage.md)
- [Database Setup](docs/database-setup.md)
- [Supported Formats](docs/supported-formats.md)
- [Troubleshooting](docs/troubleshooting.md)

pgEdge Document Loader is a command-line tool for loading documents from various formats into PostgreSQL databases.  Full documentation is available at:
[https://pgedge.github.io/pgedge-docloader](https://pgedge.github.io/pgedge-docloader)

## Overview

The pgEdge Document Loader automatically converts documents (HTML, Markdown,
and reStructuredText) to Markdown format and loads them into a PostgreSQL
database with extracted metadata.

## Features

- **Multiple Format Support**: HTML, Markdown, and reStructuredText
    - **HTML** (`.html`, `.htm`) - Extracts title from `<title>` tag
    - **Markdown** (`.md`) - Extracts title from first `#` heading
    - **reStructuredText** (`.rst`) - Converts `.rst` to Markdown
- **Automatic Conversion**: All formats converted to Markdown
- **Metadata Extraction**: Titles, filenames, timestamps
- **Flexible Input**: Single file, directory, or glob patterns (including `**` recursive matching)
- **Database Flexibility**: Configurable column mappings
- **Custom Metadata Columns**: Add fixed values to custom columns for every row
- **Update Mode**: Update existing rows or insert new ones
- **Transactional**: All-or-nothing processing with automatic rollback
- **Secure**: Password from environment, .pgpass, or interactive prompt
- **Configuration Files**: Reusable YAML configuration

## Prerequisites

Before installing and using pgEdge Document Loader, download and install:

- Go 1.21 or later
- PostgreSQL 12 or later

## Quick Start

Getting started with pgEdge Document Loader involves three steps:

1. Install the tool.
2. Create a table in your Postgres database to hold the loaded content.
3. Run the `pgedge-docloader` executable.

**Installing pgEdge Document Loader**

Use the following commands to download and build `pgedge-docloader`:

```bash
git clone https://github.com/pgedge/pgedge-docloader.git
cd pgedge-docloader
make build
make install
```

**Creating a Postgres Table**

Create a table in your Postgres database that has the [appropriate columns](https://github.com/pgEdge/pgedge-docloader/blob/main/docs/configuration.md#column-mappings) to hold the extracted documentation content:

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

**Invoking pgedge-docloader**

When invoking `pgedge-docloader`, you can [specify preferences on the command line](#command-line-options), or with a configuration file.  Use the following form on the command line:

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

To manage deployment preferences in a [configuration file](https://github.com/pgEdge/pgedge-docloader/blob/main/docs/configuration.md#configuration), save your deployment details in a file, and then include the `--config` keyword when invoking `pgedge-docloader`:

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

# Run with a config file
export PGPASSWORD=mypassword
pgedge-docloader --config config.yml
```

## Command-Line Options

When invoking `pgedge-docloader` on the command line, you can include the following options:

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

To load content from a specified directory (./documentation):

```bash
pgedge-docloader --source ./documentation --config config.yml
```

To load content that matches a specified pattern (`"./docs/**/*.md"`):

```bash
pgedge-docloader --source "./docs/**/*.md" --config config.yml
```
Include ** to enforce recursive matching across all subdirectories; for example:

* `docs/**/*.md` - matches all .md files in docs and all subdirectories.
* `docs/*.md` - matches only .md files directly in docs (not subdirectories).

To insert new rows and update existing rows (keeping your table in sync with documentation updates), include the following command syntax:

```bash
pgedge-docloader --source ./docs --config config.yml --update
```

## Support

To review documentation or to open an issue, visit:

- Documentation: [https://pgedge.github.io/pgedge-docloader](https://pgedge.github.io/pgedge-docloader)
- Issues: [GitHub Issues](https://github.com/pgedge/pgedge-docloader/issues)

## Development

Use the following command to create the binary:

```bash
make build
```

The `make` command creates the binary at `bin/pgedge-docloader`. To test locally without installing:

```bash
./bin/pgedge-docloader --help
```

To run project tests, use the following command:

```bash
make test
```

To run the `golangci-lint` linter:

```bash
make lint
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

This project is licensed under the PostgreSQL License. See [LICENCE.md](LICENCE.md) for details.

## Project Status

This project is under active development. See the documentation for the latest features and updates.

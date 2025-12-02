# pgEdge Document Loader

## Overview

The pgEdge Document Loader is a command-line tool written in Go that loads
documents from various formats into a PostgreSQL database. The tool
automatically converts documents to Markdown format and extracts metadata
before storing them in the database.

## Supported Formats

The tool supports the following document formats:

- **HTML** (`.html`, `.htm`) - Extracts title from `<title>` tag
- **Markdown** (`.md`) - Extracts title from first `#` heading
- **reStructuredText** (`.rst`) - Extracts title from underlined headings

## Key Features

- Automatic document format detection
- Conversion to Markdown format
- Metadata extraction (title, filename, timestamps)
- Flexible column mapping
- Support for single files, directories, and glob patterns
- Update or insert mode (upsert functionality)
- Transaction-based processing with automatic rollback on errors
- Configuration file support for reusable setups
- Secure password handling (environment variable, .pgpass, or interactive)

## Quick Start

1. Install the tool:
   ```bash
   make install
   ```

2. Create a database table (see [Database Setup](database-setup.md))

3. Run the tool:
   ```bash
   pgedge-docloader \
     --source ./docs \
     --db-host localhost \
     --db-name mydb \
     --db-user myuser \
     --db-table documents \
     --col-doc-content content \
     --col-file-name filename
   ```

## Documentation

- [Installation](installation.md)
- [Configuration](configuration.md)
- [Usage](usage.md)
- [Database Setup](database-setup.md)
- [Supported Formats](supported-formats.md)
- [Troubleshooting](troubleshooting.md)

## License

This project is licensed under the PostgreSQL License. See
[LICENCE.md](https://github.com/pgEdge/pgedge-docloader/blob/main/LICENCE.md) for details.

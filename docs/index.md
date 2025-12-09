# pgEdge Document Loader

The pgEdge Document Loader is a command-line tool written in Go that loads
documents from various formats into a Postgres database. The tool
automatically converts documents to Markdown format and extracts metadata
before storing them in the database.

pgEdge Document Loader supports the following document formats:

- **HTML** (`.html`, `.htm`) - Extracts the document title from `<title>` tags.
- **Markdown** (`.md`) - Extracts the title from first `#` headings.
- **reStructuredText** (`.rst`) - Extracts the title from underlined headings.
- **DocBook SGML/XML** (`.sqml`, `.xml` ) - Extracts the title from `<title>` or `<refentrytitle>` tags (PostgreSQL-style reference pages use `<refentrytitle>`).

**Key Features**

The Document Loader supports:

- automatic document format detection.
- conversion to Markdown format.
- metadata extraction (title, filename, timestamps).
- flexible column mapping.
- importing from single files, directories, and user-specified glob patterns.
- documentation updates (upsert functionality) in update or insert mode.
- transaction-based processing with automatic rollback on errors.
- configuration file storage of execution details for reusable setups.
- secure password handling (environment variable, .pgpass, or interactive).

**License**

This project is licensed under the [PostgreSQL License](LICENCE.md).

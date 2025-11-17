# Database Setup

This guide shows how to set up your PostgreSQL database for use with the
pgEdge Document Loader.

## Table Requirements

The tool can work with any table structure, as long as you map the
appropriate columns. However, you must create the table before running the
tool.

## Example Table Schemas

### Minimal Schema

A minimal table with just content and filename:

```sql
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    content TEXT,
    filename TEXT UNIQUE
);
```

Use with:

```bash
pgedge-docloader \
  --source ./docs \
  --db-table documents \
  --col-doc-content content \
  --col-file-name filename \
  ... other connection options ...
```

### Recommended Schema

A recommended schema with full metadata:

```sql
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    title TEXT,
    content TEXT NOT NULL,
    source BYTEA,
    filename TEXT UNIQUE NOT NULL,
    file_modified TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups
CREATE INDEX idx_documents_filename ON documents(filename);

-- Index for full-text search (optional)
CREATE INDEX idx_documents_content_fts ON documents
    USING gin(to_tsvector('english', content));
```

Use with:

```bash
pgedge-docloader \
  --source ./docs \
  --db-table documents \
  --col-doc-title title \
  --col-doc-content content \
  --col-source-content source \
  --col-file-name filename \
  --col-file-modified file_modified \
  --col-row-created created_at \
  --col-row-updated updated_at \
  ... other connection options ...
```

### Schema for Vector Search

For use with pgvector (semantic search):

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    title TEXT,
    content TEXT NOT NULL,
    source BYTEA,
    filename TEXT UNIQUE NOT NULL,
    file_modified TIMESTAMP,
    embedding vector(1536),  -- For OpenAI embeddings
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for vector similarity search
CREATE INDEX idx_documents_embedding ON documents
    USING ivfflat (embedding vector_cosine_ops);

-- Index for filename lookups
CREATE INDEX idx_documents_filename ON documents(filename);
```

Note: The embedding column must be populated separately using an embedding
model.

## Column Data Types

The tool expects the following data types for each column type:

- **doc_title**: `TEXT` or `VARCHAR`
- **doc_content**: `TEXT` or `VARCHAR`
- **source_content**: `BYTEA` (binary data for storing original source)
- **file_name**: `TEXT` or `VARCHAR` (recommend UNIQUE constraint for update
  mode)
- **file_created**: `TIMESTAMP` or `TIMESTAMPTZ`
- **file_modified**: `TIMESTAMP` or `TIMESTAMPTZ`
- **row_created**: `TIMESTAMP` or `TIMESTAMPTZ` (recommend DEFAULT
  CURRENT_TIMESTAMP)
- **row_updated**: `TIMESTAMP` or `TIMESTAMPTZ` (recommend DEFAULT
  CURRENT_TIMESTAMP)

## Update Mode Considerations

When using `--update` mode:

1. The `filename` column should have a UNIQUE constraint
2. The tool matches existing rows by filename
3. If a match is found, the row is updated
4. If no match is found, a new row is inserted

Example schema for update mode:

```sql
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    filename TEXT UNIQUE NOT NULL,  -- UNIQUE constraint required
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Permissions

The database user must have the following permissions:

```sql
-- Grant INSERT permission
GRANT INSERT ON documents TO myuser;

-- Grant UPDATE permission (for --update mode)
GRANT UPDATE ON documents TO myuser;

-- Grant SELECT permission (for checking existing rows)
GRANT SELECT ON documents TO myuser;

-- Grant USAGE on sequence (for SERIAL columns)
GRANT USAGE, SELECT ON SEQUENCE documents_id_seq TO myuser;
```

## Creating the Database

Complete example of setting up a new database:

```sql
-- Connect as superuser
psql -U postgres

-- Create database
CREATE DATABASE docdb;

-- Create user
CREATE USER docloader WITH PASSWORD 'secure_password';

-- Connect to the new database
\c docdb

-- Create table
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    title TEXT,
    content TEXT NOT NULL,
    source BYTEA,
    filename TEXT UNIQUE NOT NULL,
    file_modified TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE ON documents TO docloader;
GRANT USAGE, SELECT ON SEQUENCE documents_id_seq TO docloader;

-- Create indexes
CREATE INDEX idx_documents_filename ON documents(filename);
CREATE INDEX idx_documents_content_fts ON documents
    USING gin(to_tsvector('english', content));
```

## Verification

Verify your setup:

```sql
-- Check table structure
\d documents

-- Check permissions
SELECT grantee, privilege_type
FROM information_schema.role_table_grants
WHERE table_name = 'documents';

-- Test insert (as the docloader user)
INSERT INTO documents (title, content, filename)
VALUES ('Test', 'Test content', 'test.md');

-- Verify
SELECT * FROM documents WHERE filename = 'test.md';

-- Clean up test data
DELETE FROM documents WHERE filename = 'test.md';
```

## Common Table Patterns

### Simple Documentation Store

```sql
CREATE TABLE docs (
    id SERIAL PRIMARY KEY,
    title TEXT,
    body TEXT,
    path TEXT UNIQUE,
    updated TIMESTAMP
);
```

### Knowledge Base with Categories

```sql
CREATE TABLE knowledge_base (
    id SERIAL PRIMARY KEY,
    category TEXT,
    title TEXT,
    content_md TEXT,
    content_original BYTEA,
    source_file TEXT UNIQUE,
    modified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_kb_category ON knowledge_base(category);
```

### Multi-language Documentation

```sql
CREATE TABLE documentation (
    id SERIAL PRIMARY KEY,
    language TEXT DEFAULT 'en',
    title TEXT,
    content TEXT,
    filepath TEXT NOT NULL,
    modified TIMESTAMP,
    UNIQUE(language, filepath)
);

CREATE INDEX idx_docs_language ON documentation(language);
```

## Next Steps

- [Usage](usage.md) - Learn how to run the tool
- [Configuration](configuration.md) - Set up your configuration file

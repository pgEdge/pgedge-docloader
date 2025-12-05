# Configuring the Postgres Database

Before invoking pgEdge Document Loader, you need to install Postgres and create a table that will hold the document contents.  The tool can work with any table structure, as long as you map the columns appropriately.

This page will walk you through configuring your PostgreSQL database for use with the pgEdge Document Loader.  This page assumes you have installed Postgres version 12 or later.

Use the following commands to configure your database:

```sql
-- Connect as a superuser
psql -U postgres

-- Create the database
CREATE DATABASE docdb;

-- Create a user
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

After configuring the database and creating a table, you can verify your setup with the following commands:

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

The following command snippet references the table:

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

## Creating a Table for Vector Searches

The following commands create the vector extension, a table for use with pgvector (semantic search), and indexes:

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

Note: The embedding column must be populated separately using an embedding model.


## Examples - Common Table Patterns

The following examples demonstrate some useful table configurations.

**Simple Documentation Store**

```sql
CREATE TABLE docs (
    id SERIAL PRIMARY KEY,
    title TEXT,
    body TEXT,
    path TEXT UNIQUE,
    updated TIMESTAMP
);
```

**Knowledge Base with Categories**

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

**Multi-language Documentation**

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
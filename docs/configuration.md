# Configuration

The pgEdge Document Loader can be configured using a YAML configuration file
and/or command-line flags. Command-line flags always take precedence over
configuration file settings.

## Configuration File

Create a YAML configuration file (e.g., `config.yml`):

```yaml
# Source documents
source: "./docs"
strip-path: false

# Database connection
db-host: localhost
db-port: 5432
db-name: mydb
db-user: myuser
db-sslmode: prefer
db-table: documents

# SSL/TLS certificates (optional)
db-sslcert: /path/to/client-cert.pem
db-sslkey: /path/to/client-key.pem
db-sslrootcert: /path/to/ca-cert.pem

# Column mappings
col-doc-title: title
col-doc-content: content
col-source-content: source
col-file-name: filename
col-file-created: created
col-file-modified: modified
col-row-created: created_at
col-row-updated: updated_at

# Operation mode
update: true
```

Use the configuration file:

```bash
pgedge-docloader --config config.yml
```

## Configuration Options

### Source Options

- **source** (required): Path to file, directory, or glob pattern
- **strip-path**: Remove directory path from filenames (default: false)

### Database Connection

- **db-host**: Database hostname (default: localhost)
- **db-port**: Database port (default: 5432)
- **db-name** (required): Database name
- **db-user** (required): Database username
- **db-sslmode**: SSL mode - disable, allow, prefer, require, verify-ca,
  verify-full (default: prefer)
- **db-table** (required): Target table name

### SSL/TLS Configuration

- **db-sslcert**: Path to client SSL certificate
- **db-sslkey**: Path to client SSL key
- **db-sslrootcert**: Path to SSL root certificate

### Column Mappings

Map document data to table columns. At least one column must be specified:

- **col-doc-title**: Column for document title (TEXT)
- **col-doc-content**: Column for converted Markdown content (TEXT)
- **col-source-content**: Column for original source (BYTEA)
- **col-file-name**: Column for filename (TEXT)
- **col-file-created**: Column for file creation timestamp (TIMESTAMP)
- **col-file-modified**: Column for file modification timestamp (TIMESTAMP)
- **col-row-created**: Column for row creation timestamp (TIMESTAMP)
- **col-row-updated**: Column for row update timestamp (TIMESTAMP)

### Custom Metadata Columns

Add fixed values to custom columns for every row inserted. This is useful
for storing multiple documentation sets in a single table with distinguishing
metadata.

**Configuration file format:**

```yaml
custom-columns:
  product: "pgAdmin 4"
  version: "v9.9"
  environment: "production"
```

**Command-line format:**

```bash
pgedge-docloader --set-column product="pgAdmin 4" --set-column version="v9.9"
```

The `--set-column` flag can be specified multiple times. Command-line values
override config file values for the same column name.

### Operation Mode

- **update**: Enable update mode - update existing rows (matched by
  filename) or insert new ones (default: false)

## Password Configuration

Database passwords are never stored in configuration files. The tool obtains
passwords in this order of priority:

1. **PGPASSWORD environment variable**

   ```bash
   export PGPASSWORD=mypassword
   pgedge-docloader --config config.yml
   ```

2. **~/.pgpass file** (format: `hostname:port:database:username:password`)

   ```
   localhost:5432:mydb:myuser:mypassword
   ```

   Ensure proper permissions:

   ```bash
   chmod 600 ~/.pgpass
   ```

3. **Passwordless authentication** - If no password is found, the tool will
   attempt to connect without one, allowing PostgreSQL to use configured
   authentication methods such as:
   - Trust authentication
   - Peer authentication
   - Certificate-based authentication (using `db-sslcert` and `db-sslkey`)

If a password is required but not provided through PGPASSWORD or .pgpass,
PostgreSQL will return an authentication error with a clear message.

## Path Resolution

When using a configuration file, relative paths are resolved relative to the
configuration file's directory. For example:

```yaml
source: ../docs              # Relative to config file
db-sslcert: ./certs/client.pem  # Relative to config file
```

Without a configuration file, relative paths are relative to the current
working directory.

## Command-Line Flags

All configuration options have corresponding command-line flags. Use
`--help` to see all available flags:

```bash
pgedge-docloader --help
```

Command-line flags override configuration file values.

## Example Configurations

### Minimal Configuration

```yaml
source: "./docs/*.md"
db-host: localhost
db-name: mydb
db-user: myuser
db-table: documents
col-doc-content: content
col-file-name: filename
```

### Full Configuration

```yaml
source: "./documentation"
strip-path: true

db-host: db.example.com
db-port: 5432
db-name: production_db
db-user: doc_loader
db-sslmode: verify-full
db-sslcert: ./certs/client.pem
db-sslkey: ./certs/client-key.pem
db-sslrootcert: ./certs/ca.pem
db-table: knowledge_base

col-doc-title: title
col-doc-content: content_markdown
col-source-content: content_original
col-file-name: source_file
col-file-modified: file_modified_at
col-row-created: created_at
col-row-updated: updated_at

custom-columns:
  product: "pgAdmin 4"
  version: "v9.9"

update: true
```

## Next Steps

- [Usage](usage.md) - Learn how to run the tool
- [Database Setup](database-setup.md) - Create appropriate database tables

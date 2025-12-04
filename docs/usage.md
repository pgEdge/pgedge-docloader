# Usage

This guide covers common usage patterns for the pgEdge Document Loader.

## Basic Usage

### Load a Single File

```bash
pgedge-docloader \
  --source /path/to/document.md \
  --db-host localhost \
  --db-name mydb \
  --db-user myuser \
  --db-table documents \
  --col-doc-content content \
  --col-file-name filename
```

### Load a Directory

Load all supported files in a directory:

```bash
pgedge-docloader \
  --source /path/to/docs \
  --db-host localhost \
  --db-name mydb \
  --db-user myuser \
  --db-table documents \
  --col-doc-content content \
  --col-file-name filename
```

### Load with Glob Pattern

Load files matching a pattern. Use `**` for recursive matching across all
subdirectories:

```bash
pgedge-docloader \
  --source "/path/to/docs/**/*.md" \
  --db-host localhost \
  --db-name mydb \
  --db-user myuser \
  --db-table documents \
  --col-doc-content content \
  --col-file-name filename
```

The `**` pattern recursively matches all subdirectories. For example:

- `docs/**/*.md` - All .md files in docs and all subdirectories
- `docs/*.md` - Only .md files directly in docs (not subdirectories)

## Using a Configuration File

Create a configuration file and use it:

```bash
pgedge-docloader --config config.yml
```

Override config file settings with command-line flags:

```bash
pgedge-docloader --config config.yml --source /different/path
```

## Update Mode

In update mode, the tool updates existing rows (matched by filename) or
inserts new ones:

```bash
pgedge-docloader \
  --config config.yml \
  --update
```

This is useful for keeping the database in sync with document changes.

## Stripping Paths

By default, the full path is stored in the filename column. Use
`--strip-path` to store only the base filename:

```bash
pgedge-docloader \
  --source /long/path/to/docs \
  --strip-path \
  --db-host localhost \
  --db-name mydb \
  --db-user myuser \
  --db-table documents \
  --col-file-name filename
```

With `--strip-path`, `/long/path/to/docs/file.md` becomes `file.md`.

## Column Mapping

Map document data to different table columns:

```bash
pgedge-docloader \
  --source ./docs \
  --db-host localhost \
  --db-name mydb \
  --db-user myuser \
  --db-table documents \
  --col-doc-title title \
  --col-doc-content content \
  --col-source-content original \
  --col-file-name filename \
  --col-file-modified modified_at \
  --col-row-created created_at \
  --col-row-updated updated_at
```

You can map any combination of columns. The tool will only populate the
columns you specify.

## Custom Metadata Columns

Add fixed values to custom columns for every row inserted. This is useful for
storing multiple documentation sets in a single table:

**Using command-line flags:**

```bash
pgedge-docloader \
  --source ./docs/pgadmin \
  --config base-config.yml \
  --set-column product="pgAdmin 4" \
  --set-column version="v9.9" \
  --set-column environment="production"
```

**Using configuration file:**

```yaml
source: "./docs/pgadmin"
db-host: localhost
db-name: docdb
db-user: docuser
db-table: all_docs
col-doc-content: content
col-file-name: filename

custom-columns:
  product: "pgAdmin 4"
  version: "v9.9"
  environment: "production"
```

The `--set-column` flag can be specified multiple times. Command-line values
override config file values for the same column name.

### Example: Multiple Documentation Sets

Store documentation for different products in the same table:

```bash
# Load pgAdmin documentation
pgedge-docloader \
  --source ./docs/pgadmin \
  --config base-config.yml \
  --set-column product="pgAdmin 4" \
  --set-column version="v9.9"

# Load pgEdge documentation
pgedge-docloader \
  --source ./docs/pgedge \
  --config base-config.yml \
  --set-column product="pgEdge" \
  --set-column version="v2.5"
```

Then query by product:

```sql
SELECT title, content FROM all_docs WHERE product = 'pgAdmin 4';
```

## SSL/TLS Connections

Connect using SSL/TLS with client certificates:

```bash
pgedge-docloader \
  --source ./docs \
  --db-host secure.example.com \
  --db-name mydb \
  --db-user myuser \
  --db-table documents \
  --db-sslmode verify-full \
  --db-sslcert ./certs/client.pem \
  --db-sslkey ./certs/client-key.pem \
  --db-sslrootcert ./certs/ca.pem \
  --col-doc-content content \
  --col-file-name filename
```

SSL modes:

- `disable` - No SSL
- `allow` - Try SSL, fall back to non-SSL
- `prefer` - Try SSL, fall back to non-SSL (default)
- `require` - Require SSL, but don't verify certificates
- `verify-ca` - Require SSL and verify CA certificate
- `verify-full` - Require SSL and verify certificate and hostname

## Password Options

### Using Environment Variable

```bash
export PGPASSWORD=mypassword
pgedge-docloader --config config.yml
```

### Using .pgpass File

Create `~/.pgpass`:

```
localhost:5432:mydb:myuser:mypassword
```

Set permissions:

```bash
chmod 600 ~/.pgpass
```

### Interactive Prompt

If no password is found, the tool will prompt:

```bash
pgedge-docloader --config config.yml
Enter database password: ****
```

## Viewing Help

Get help on all command-line options:

```bash
pgedge-docloader --help
```

Get version information:

```bash
pgedge-docloader version
```

List supported document formats:

```bash
pgedge-docloader formats
```

## Processing Summary

After processing, the tool displays a summary:

```
Processing files from: ./docs
Processed 15 file(s), skipped 2 file(s)
Connecting to database myuser@localhost:5432/mydb

=== Processing Summary ===
Files processed: 15
Files skipped:   2
Rows inserted:   15
Rows updated:    0
=========================
```

## Error Handling

If any error occurs during processing or database operations:

- All database changes are rolled back (nothing is committed)
- The tool exits with a non-zero status code
- A detailed error message is displayed

Example:

```
Error: failed to insert documents: pq: duplicate key value violates
unique constraint "documents_filename_key"
```

## Advanced Examples

### Load Only Markdown Files

```bash
pgedge-docloader --source "./docs/*.md" --config config.yml
```

### Load with Full Metadata

```bash
pgedge-docloader \
  --source ./docs \
  --db-host localhost \
  --db-name mydb \
  --db-user myuser \
  --db-table knowledge_base \
  --col-doc-title title \
  --col-doc-content content_markdown \
  --col-source-content content_original \
  --col-file-name source_file \
  --col-file-modified file_modified_at \
  --col-row-created created_at \
  --col-row-updated updated_at \
  --update
```

### Automated Sync with Cron

Add to crontab for regular updates:

```cron
# Sync documentation every hour
0 * * * * /usr/local/bin/pgedge-docloader --config /etc/docloader/config.yml --update
```
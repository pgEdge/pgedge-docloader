# Configuration

The pgEdge Document Loader can be deployed with preferences saved in a [YAML configuration file](#specifying-options-in-a-configuration-file) and/or [command-line flags](#specifying-options-on-the-command-line). 

!!! note

    Command-line flags always take precedence over configuration file settings.


## Specifying Options in a Configuration File

To save your deployment preferences in a file, create a YAML-formatted configuration file (for example, `config.yml`):

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

Then, when you invoke `pgedge-docloader`, include the `--config` flag and the configuration file name:

```bash
pgedge-docloader --config config.yml
```

### Configuration File Options

Use the following options to specify details about the source document:

| Option      | Required | Description                                         | Default |
|-------------|----------|-----------------------------------------------------|---------|
| source      | Yes      | Path to file, directory, or glob pattern            | —       |
| strip-path  | No       | Remove directory path from filenames                | false   |

Use the following options in a configuration file to specify details about the database connection:

| Option     | Required | Description                                                       | Default     |
|------------|----------|-------------------------------------------------------------------|-------------|
| db-host    | No       | Database hostname                                                 | localhost   |
| db-port    | No       | Database port                                                     | 5432        |
| db-name    | Yes      | Database name                                                     | —           |
| db-user    | Yes      | Database username                                                 | —           |
| db-sslmode | No       | SSL mode (disable, allow, prefer, require, verify-ca, verify-full)| prefer      |
| db-table   | Yes      | Target table name                                                 | —           |

Use the following options to specify details about the SSL/TLS configuration:

| Option         | Required | Description                         | Default |
|----------------|----------|-------------------------------------|---------|
| db-sslcert     | No       | Path to client SSL certificate      | —       |
| db-sslkey      | No       | Path to client SSL key              | —       |
| db-sslrootcert | No       | Path to SSL root certificate        | —       |

Use the following options to specify details about column mappings:

| Option             | Required | Description                                            | Default |
|--------------------|----------|--------------------------------------------------------|---------|
| col-doc-title      | No       | Column for document title (TEXT)                       | —       |
| col-doc-content    | No       | Column for converted Markdown content (TEXT)           | —       |
| col-source-content | No       | Column for original source (BYTEA)                     | —       |
| col-file-name      | No       | Column for filename (TEXT)                             | —       |
| col-file-created   | No       | Column for file creation timestamp (TIMESTAMP)         | —       |
| col-file-modified  | No       | Column for file modification timestamp (TIMESTAMP)     | —       |
| col-row-created    | No       | Column for row creation timestamp (TIMESTAMP)          | —       |
| col-row-updated    | No       | Column for row update timestamp (TIMESTAMP)            | —       |


## Specifying Options on the Command-Line

All configuration options have corresponding command-line flags. Use `--help` to see all available flags:

```bash
pgedge-docloader --help
```

The following command demonstrates specifying options on the command line; in the command, each command line option is followed by the column name in which the content will be stored:

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

Command-line flags override configuration file values.

FIXME - Add table of --help options with field names?


## Using Custom Metadata Columns

You can add fixed values to custom columns for each row inserted. This is useful for storing multiple documentation sets in a single table with distinguishing metadata.

Within the configuration file, use the following format:

```yaml
custom-columns:
  product: "pgAdmin 4"
  version: "v9.9"
  environment: "production"
```

To specify metadata columns on the command line, include the following options:

```bash
pgedge-docloader --set-column product="pgAdmin 4" --set-column version="v9.9"
```

You can specify the `--set-column` flag multiple times. Command-line values override configuration file values for the same column name.


## Examples

The following options specify the minimal configuration required by Document Loader:

```yaml
source: "./docs/*.md"
db-host: localhost
db-name: mydb
db-user: myuser
db-table: documents
col-doc-content: content
col-file-name: filename
```

The following options specify a complete configuration:

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
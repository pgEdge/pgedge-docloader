# pgEdge Document Loader

[![CI](https://github.com/pgEdge/pgedge-docloader/actions/workflows/ci.yml/badge.svg)](https://github.com/pgEdge/pgedge-docloader/actions/workflows/ci.yml)

  - [Introduction](docs/index.md)
      - [Best Practices](docs/best_practices.md)
  - Installing pgEdge Document Loader
      - [Configuring the Postgres Database](docs/database-setup.md)
      - [Installing Document Loader](docs/installation.md)
      - [Document Loader Configuration](docs/configuration.md)
      - [pgEdge Document Loader Quickstart](docs/quickstart.md)
  - Using pgEdge Document Loader
      - [Using Document Loader](docs/usage.md)
      - [Using Custom Metadata Columns](docs/metadata.md)
      - [Updating a Document](docs/updating.md)
      - [Managing Authentication](docs/authentication.md)
  - Supported Formats
      - [Supported vs. Unsupported Formats](docs/unsupported-formats.md)
      - [HTML or HTM](docs/html.md)
      - [Markdown](docs/markdown.md)
      - [RST](docs/rst.md)
      - [SGML](docs/sgml.md)
  - [Troubleshooting](docs/troubleshooting.md)
  - [Licence](docs/LICENCE.md)

pgEdge Document Loader is a command-line tool for loading documents from various formats into PostgreSQL databases.  Full documentation is available [here](https://docs.pgedge.com/pgedge-docloader/).

The pgEdge Document Loader automatically converts documents (HTML, Markdown, reStructuredText, and SGML/DocBook) to Markdown format and loads them into a PostgreSQL database with extracted metadata.

**Features**

The pgEdge Document Loader automatically converts documents (HTML, Markdown, reStructuredText, and DocBook SGML/XML) to Markdown format and loads them into a PostgreSQL database with extracted metadata.

**Features**

- **Multiple Format Support**: HTML, Markdown, reStructuredText, and DocBook SGML/XML
- **Automatic Conversion**: All formats converted to Markdown
- **Metadata Extraction**: Titles, filenames, timestamps
- **Flexible Input**: Single file, directory, or glob patterns (including `**` recursive matching)
- **Database Flexibility**: Configurable column mappings
- **Custom Metadata Columns**: Add fixed values to custom columns for every row
- **Update Mode**: Update existing rows or insert new ones
- **Transactional**: All-or-nothing processing with automatic rollback
- **Secure**: Password from environment, .pgpass, or interactive prompt
- **Configuration Files**: Reusable YAML configuration

## Document Loader Quickstart

Before installing and using pgEdge Document Loader, download and install:

- Go 1.23 or later
- PostgreSQL 14 or later

Getting started with pgEdge Document Loader involves three steps:

1. Install the tool.
2. Create a table in your Postgres database to hold the loaded content.
3. Run the `pgedge-docloader` executable.

**Installing pgEdge Document Loader**

Use the following commands to [download and build `pgedge-docloader`](/docs/installation.md):

```bash
git clone https://github.com/pgedge/pgedge-docloader.git
cd pgedge-docloader
make build
make install
```

**Creating a Postgres Table**

Before invoking Document Loader, you must configure a Postgres database and create a table with the [appropriate columns](/docs/database-setup.md) to hold the extracted documentation content:

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

When invoking `pgedge-docloader`, you can [specify configuration preferences on the command line](/docs/configuration.md#specifying-options-on-the-command-line), or with a [configuration file](/docs/configuration.md#specifying-options-in-a-configuration-file).

The following command [invokes Document Loader on the command line](/docs/usage.md):

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

To manage deployment preferences in a [configuration file](/docs/configuration.md#specifying-options-in-a-configuration-file), save your deployment details in a file, and then include the `--config` keyword when invoking `pgedge-docloader`:

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

# Run with a configuration file
export PGPASSWORD=mypassword
pgedge-docloader --config config.yml
```

For a comprehensive Quickstart Guide, visit [here](/docs/quickstart.md).

## Developer Notes

This project is under active development. See the documentation for the latest
features and updates.

The pgEdge Document Loader Makefile includes clauses that run test cases or invoke the go linter.  Use the following commands:

**Running Tests**

```bash
make test
```

**Linting**

```bash
make lint
```

Your contributions are welcome! Please feel free to submit issues and pull requests.

## Support

- Documentation: [pgEdge Docloader](https://docs.pgedge.com/pgedge-docloader/)
- Issues: [GitHub Issues](https://github.com/pgedge/pgedge-docloader/issues)

## License

This project is licensed under the [PostgreSQL License](LICENCE.md).

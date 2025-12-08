# pgEdge Document Loader Tutorial

Before installing and using pgEdge Document Loader, download and install:

- Go, version 1.21 or later
- Postgres, version 12 or later

Getting started with pgEdge Document Loader involves three steps:

1. Install pgEdge Document Loader.
2. Create a table in your Postgres database to hold the loaded content.
3. Run the `pgedge-docloader` executable.

!!! hint

    The Postgres table used to store the loaded content could potentially grow to a considerable size; you should ensure that the table is stored in a location with sufficient space.

**Installing pgEdge Document Loader**

Use the following commands to clone the pgEdge Document Loader repository and build `pgedge-docloader`:

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

When invoking `pgedge-docloader`, you can [specify preferences on the command line](configuration.md), or with a configuration file.  Use the following form on the command line:

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

To manage deployment preferences in a [configuration file](configuration.md)), save your deployment details in a file, and then include the `--config` keyword when invoking `pgedge-docloader`:

```bash
# Create a config.yml file
source: ./docs
db-host: localhost
db-name: mydb
db-user: myuser
db-table: documents
col-doc-content: content
col-file-name: filename
update: true

# Run with a configuration file
export PGPASSWORD=mypassword
pgedge-docloader --config config.yml
```

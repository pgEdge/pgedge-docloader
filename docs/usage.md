# Usage

This guide covers common usage patterns for the pgEdge Document Loader.  To review online help, use the command:

```bash
pgedge-docloader --help
```

## Specifying Path Values

By default, the full path is stored in the filename column. When using a configuration file, paths are resolved relative to the configuration file's directory. For example:

```yaml
source: ../docs              # Relative to config file
db-sslcert: ./certs/client.pem  # Relative to config file
```

When invoked without a configuration file, paths are relative to the current working directory.  You can include the `--strip-path` command option on the command line to store only the base filename in your Postgres table:

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

Including the `--strip-path` option, instructs Document Loader to save `/long/path/to/docs/file_name.md` as `file_name.md`.

!!! note

    Command-line flags always take precedence over configuration file settings.


## Using pgEdge Document Loader

**Loading a Single Document**

The following command demonstrates loading a single document:

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

**Loading all of the Documents in a Directory**

The following command loads all of the supported documents in a directory:

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

**Loading Documents that Match a Specified Pattern**

The following command loads all documents that match a specified pattern. Use `**` for recursive matching across all subdirectories:

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

For example, the following command loads all Markdown documents found in the `docs` subdirectory using the configuration preferences specified in the `config.yml` file:

```bash
pgedge-docloader --source "./docs/*.md" --config config.yml
```

**Saving Multiple Documents in a Single Table**

The following commands store documentation for multiple products in the same table:

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

Then, when you query the table, you can specify a value for the `product` column to retrieve content from `pgAdmin 4`:

```sql
SELECT title, content FROM all_docs WHERE product = 'pgAdmin 4';
```

**Using Document Loader with a Configuration File**

The following command invokes Document Loader while specifying preferences in a [configuration file](configuration.md) named `config.yml`:

```bash
pgedge-docloader --config config.yml
```

You can override configuration file settings with command-line flags; command-line preferences always take precedence over options specified in a configuration file:

```bash
pgedge-docloader --config config.yml --source /different/path
```

You can map any combination of columns. The tool will only populate the columns you specify.


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

- All database changes are rolled back (nothing is committed).
- The tool exits with a non-zero status code.
- A detailed error message is displayed.

For example:

```
Error: failed to insert documents: pq: duplicate key value violates
unique constraint "documents_filename_key"
```

## Automated Sync with Cron

You can add pgEdge Document Loader to `crontab` to perform regular updates.  For example:

```cron
# Sync documentation every hour
0 * * * * /usr/local/bin/pgedge-docloader --config /etc/docloader/config.yml --update
```
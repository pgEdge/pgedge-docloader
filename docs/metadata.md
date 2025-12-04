# Adding Custom Metadata Columns

You can use metadata columns to add fixed values to custom columns for each row inserted. This is useful for storing multiple documentation sets in a single table.

The following example demonstrates using command-line options to add metadata columns:

```bash
pgedge-docloader \
  --source ./docs/pgadmin \
  --config base-config.yml \
  --set-column product="pgAdmin 4" \
  --set-column version="v9.9" \
  --set-column environment="production"
```

You can also include metadata options to include the column specifications in a configuration file:

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

You can specify the `--set-column` flag multiple times. Command-line values override configuration file values for the same column name.


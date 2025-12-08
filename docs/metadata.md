# Adding Custom Metadata Columns

You can use metadata columns to add fixed values to custom columns for each row inserted. This is useful for storing multiple documentation sets in a single table.  Within a configuration file, use the following format to specify custom columns:

```yaml
custom-columns:
  product: "pgAdmin 4"
  version: "v9.9"
  environment: "production"
```

To specify metadata columns on the command line, include the options in the following form:

```bash
pgedge-docloader --set-column product="pgAdmin 4" --set-column version="v9.9" --set-column environment="production"
```

Note that you can specify the `--set-column` flag multiple times. If you specify the same target column name in both the configuration file and the command line, `pgedge-docloader` uses the command-line value and ignores the configuration file setting.  For example, if your configuration file specifies:

custom-columns:
  product: "pgAdmin 4"
  version: "v9.9"
  environment: "production"

Then, on the command line you specify:

```bash
pgedge-docloader \
  --source ./docs/pgadmin \
  --config base-config.yml \
  --set-column product="pgAdmin 4" \
  --set-column version="v10.0" \
  --set-column environment="development"
```

The values specified for the `version` column and the `environment` column on the command line will override the values specified in the configuration file and be written to your metadata tags. 
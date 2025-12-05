# Updating a Stored Document

When invoked with the `--update` flag, the tool reviews previously stored documents and updates existing rows (matched by filename) or inserts new ones.  When using `--update` mode:

* The `filename` column should have a `UNIQUE` constraint
* The tool matches existing rows by filename
* If a match is found, the row is updated
* If no match is found, a new row is inserted

For example, the following table is suitable for update mode:

```sql
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    filename TEXT UNIQUE NOT NULL,  -- UNIQUE constraint required
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

After creating and/or modifying a target table to include the `UNIQUE` constraint, you can specify the table name in the `config.yml` file, and invoke `pgedge-docloader` with the `--update` flag:

```bash
pgedge-docloader \
  --source ./docs \
  --db-host localhost \
  --db-name mydb \
  --db-user myuser \
  --db-table documents \
  --update
```

## Performing an Automated Sync with Cron

You can add pgEdge Document Loader to `crontab` to perform regular updates.  For example:

```cron
# Sync documentation every hour
0 * * * * /usr/local/bin/pgedge-docloader --config /etc/docloader/config.yml --update
```
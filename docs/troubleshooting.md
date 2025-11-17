# Troubleshooting

Common issues and solutions when using the pgEdge Document Loader.

## Connection Issues

### Cannot Connect to Database

**Error:**

```
Error: failed to connect to database: failed to ping database:
connection refused
```

**Solutions:**

1. Verify PostgreSQL is running:

   ```bash
   pg_isready -h localhost -p 5432
   ```

2. Check host and port settings:

   ```bash
   pgedge-docloader --db-host localhost --db-port 5432 ...
   ```

3. Verify network connectivity:

   ```bash
   telnet localhost 5432
   ```

4. Check PostgreSQL is accepting connections (in `postgresql.conf`):

   ```
   listen_addresses = '*'
   ```

### Authentication Failed

**Error:**

```
Error: failed to connect to database: pq: password authentication
failed for user "myuser"
```

**Solutions:**

1. Verify password is correct

2. Set password using environment variable:

   ```bash
   export PGPASSWORD=your_password
   ```

3. Check `pg_hba.conf` allows the connection method

4. Verify user exists:

   ```sql
   SELECT usename FROM pg_user WHERE usename = 'myuser';
   ```

### SSL/TLS Issues

**Error:**

```
Error: failed to connect to database: pq: SSL is not enabled on the
server
```

**Solutions:**

1. Change SSL mode to `disable` if SSL is not needed:

   ```bash
   pgedge-docloader --db-sslmode disable ...
   ```

2. Or enable SSL in PostgreSQL (`postgresql.conf`):

   ```
   ssl = on
   ```

3. Verify certificate paths are correct:

   ```bash
   ls -la /path/to/cert.pem /path/to/key.pem
   ```

## File Processing Issues

### No Files Processed

**Error:**

```
Processing files from: ./docs
No documents to process.
```

**Solutions:**

1. Verify source path exists:

   ```bash
   ls -la ./docs
   ```

2. Check glob pattern syntax:

   ```bash
   pgedge-docloader --source "./docs/*.md" ...
   ```

3. Ensure files have supported extensions (`.html`, `.htm`, `.md`, `.rst`)

4. List supported formats:

   ```bash
   pgedge-docloader formats
   ```

### Unsupported File Type

**Error (single file):**

```
Error: unsupported file type: document.txt
```

**Solution:**

Convert the file to a supported format or use a different file.

**Message (directory):**

```
Skipping unsupported file: readme.txt
```

**Solution:**

This is informational. Unsupported files in directories/globs are skipped
automatically.

## Database Issues

### Table Does Not Exist

**Error:**

```
Error: failed to insert documents: pq: relation "documents" does not
exist
```

**Solutions:**

1. Create the table first (see [Database Setup](database-setup.md))

2. Verify table name is correct:

   ```sql
   \dt
   ```

3. Check schema if using non-public schema:

   ```bash
   --db-table myschema.documents
   ```

### Column Does Not Exist

**Error:**

```
Error: failed to insert documents: pq: column "content" of relation
"documents" does not exist
```

**Solutions:**

1. Verify column mapping matches table structure:

   ```sql
   \d documents
   ```

2. Update column mapping:

   ```bash
   --col-doc-content actual_column_name
   ```

### Permission Denied

**Error:**

```
Error: failed to insert documents: pq: permission denied for table
documents
```

**Solutions:**

1. Grant necessary permissions:

   ```sql
   GRANT SELECT, INSERT, UPDATE ON documents TO myuser;
   GRANT USAGE, SELECT ON SEQUENCE documents_id_seq TO myuser;
   ```

2. Verify user has permissions:

   ```sql
   SELECT grantee, privilege_type
   FROM information_schema.role_table_grants
   WHERE table_name = 'documents';
   ```

### Duplicate Key Violation

**Error:**

```
Error: failed to insert documents: pq: duplicate key value violates
unique constraint "documents_filename_key"
```

**Solutions:**

1. Use update mode to update existing rows:

   ```bash
   pgedge-docloader --update ...
   ```

2. Or delete existing rows:

   ```sql
   DELETE FROM documents WHERE filename = 'duplicate.md';
   ```

3. Or remove UNIQUE constraint if not needed

### Type Mismatch

**Error:**

```
Error: failed to insert documents: pq: column "source" is of type bytea
but expression is of type text
```

**Solutions:**

1. Verify column types match expectations (see [Database
   Setup](database-setup.md))

2. For source content, use BYTEA type:

   ```sql
   ALTER TABLE documents
   ALTER COLUMN source TYPE bytea
   USING source::bytea;
   ```

## Configuration Issues

### Config File Not Found

**Error:**

```
Error: failed to load configuration: failed to read config file: open
config.yml: no such file or directory
```

**Solutions:**

1. Verify config file path:

   ```bash
   ls -la config.yml
   ```

2. Use absolute path:

   ```bash
   pgedge-docloader --config /full/path/to/config.yml
   ```

### Invalid YAML Syntax

**Error:**

```
Error: failed to load configuration: failed to read config file: yaml:
line 5: could not find expected ':'
```

**Solutions:**

1. Validate YAML syntax:

   ```bash
   yamllint config.yml
   ```

2. Check for correct indentation (use spaces, not tabs)

3. Ensure colons are followed by spaces

### Missing Required Options

**Error:**

```
Error: failed to load configuration: source path is required
```

**Solutions:**

1. Provide required options via config file or command line

2. Required options:
   - `source`
   - `db-host`
   - `db-name`
   - `db-user`
   - `db-table`
   - At least one column mapping

## Performance Issues

### Slow Processing

If processing is slow:

1. **Database connection:** Use connection pooling (already enabled)

2. **Network latency:** Use local database or faster network connection

3. **Indexes:** Ensure appropriate indexes exist (see [Database
   Setup](database-setup.md))

### High Memory Usage

For large documents:

1. Process files in smaller batches

2. Use glob patterns to process subsets:

   ```bash
   pgedge-docloader --source "./docs/section1/*.md" ...
   pgedge-docloader --source "./docs/section2/*.md" ...
   ```

## Getting Help

If you encounter issues not covered here:

1. Check the [Usage](usage.md) guide

2. Verify your [Configuration](configuration.md)

3. Review [Database Setup](database-setup.md)

4. Run with verbose output (if available in future versions)

5. Report issues at: https://github.com/pgedge/pgedge-docloader/issues

## Common Debugging Steps

1. **Test database connection:**

   ```bash
   psql -h localhost -U myuser -d mydb
   ```

2. **Test with minimal config:**

   ```bash
   pgedge-docloader \
     --source testdata/sample.md \
     --db-host localhost \
     --db-name test \
     --db-user postgres \
     --db-table test_docs \
     --col-doc-content content \
     --col-file-name filename
   ```

3. **Check PostgreSQL logs:**

   ```bash
   tail -f /var/log/postgresql/postgresql-*.log
   ```

4. **Verify table structure:**

   ```sql
   \d+ documents
   ```

5. **Test manual insert:**

   ```sql
   INSERT INTO documents (content, filename)
   VALUES ('test', 'test.md');
   ```

## Next Steps

- [Usage](usage.md) - Review usage examples
- [Database Setup](database-setup.md) - Verify database setup
- [Configuration](configuration.md) - Review configuration options

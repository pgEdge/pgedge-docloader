# Password Options

## Specifying a Password

Database passwords are never stored in a configuration file. The tool obtains passwords in this order of priority:

1. pgEdge Document Loader first checks the `PGPASSWORD` environment variable:

   ```bash
   export PGPASSWORD=mypassword
   pgedge-docloader --config config.yml
   ```

2. It then checks the [`~/.pgpass file`](https://www.postgresql.org/docs/18/libpq-pgpass.html) for an entry:

   ```
   localhost:5432:mydb:myuser:mypassword
   ```

   Your `/.pgpass` file must have proper permissions:

   ```bash
   chmod 600 ~/.pgpass
   ```

!!! note

    If a password is required but not provided through `PGPASSWORD` or `.pgpass`, PostgreSQL will return an authentication error with a clear message.

3. If Document Loader doesn't find a password in the two previous locations, it then attempts passwordless authentication. This allows PostgreSQL to use configured authentication methods such as:

   - Trust authentication
   - Peer authentication
   - Certificate-based authentication (using `db-sslcert` and `db-sslkey`)

If no password is found and an alternative authentication method is not configured, the tool will prompt:

```bash
pgedge-docloader --config config.yml
Enter database password: ****
```

### Using an Environment Variable to Specify a Password

```bash
export PGPASSWORD=mypassword
pgedge-docloader --config config.yml
```

### Using the .pgpass File to Store a Password

Create `~/.pgpass`:

```
localhost:5432:mydb:myuser:mypassword
```

Set permissions:

```bash
chmod 600 ~/.pgpass
```

## Using an SSL/TLS Connection

Include the following options to connect using SSL/TLS with client certificates:

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

The supported SSL modes are:

- `disable` - No SSL
- `allow` - Try SSL, fall back to non-SSL
- `prefer` - Try SSL, fall back to non-SSL (default)
- `require` - Require SSL, but don't verify certificates
- `verify-ca` - Require SSL and verify CA certificate
- `verify-full` - Require SSL and verify certificate and hostname


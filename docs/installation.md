# Installation

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Make (optional, for using Makefile targets)

## Building from Source

### Clone the Repository

```bash
git clone https://github.com/pgedge/pgedge-docloader.git
cd pgedge-docloader
```

### Download Dependencies

```bash
make deps
```

Or using Go directly:

```bash
go mod download
```

### Build the Binary

```bash
make build
```

This creates the `pgedge-docloader` binary in the current directory.

### Install System-Wide

To install the binary to `/usr/local/bin`:

```bash
make install
```

To install to a custom location:

```bash
PREFIX=/opt/local make install
```

## Verify Installation

After installation, verify the tool is working:

```bash
pgedge-docloader version
```

Check supported formats:

```bash
pgedge-docloader formats
```

## Docker Installation (Future)

Docker images may be provided in future releases.

## Next Steps

- [Database Setup](database-setup.md) - Create your database table
- [Configuration](configuration.md) - Set up a configuration file
- [Usage](usage.md) - Learn how to use the tool

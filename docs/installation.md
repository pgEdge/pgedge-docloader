# Installation

pgEdge Document Loader is open-source and licensed with the [PostgreSQL license](LICENCE.md).  You can download pgEdge Document Loader source code from the [pgEdge repository](https://github.com/pgEdge/pgedge-docloader).

## Prerequisites

Before building Document Loader, install:

- Go 1.21 or later
- PostgreSQL 12 or later
- Make (optional, for using Makefile targets)

## Building from Source

To build Document Loader from source, clone the pgedge-docloader repository:

```bash
git clone https://github.com/pgedge/pgedge-docloader.git
cd pgedge-docloader
```
Then, use `make` to ensure that your Go installation is configured properly:

```bash
make deps
```

Alternatively, you can use the command:

```bash
go mod download
```

Then, use `make` to build the Document Loader binary:

```bash
make build
```

The `make build` command creates the `pgedge-docloader` binary in the current directory.  If you'd prefer to install the binary in `/usr/local/bin`, use the command:

```bash
make install
```

To install the binary in a custom location, specify the installation path with the `make` command:

```bash
PREFIX=/opt/local make install
```
**Verify the Installation**

After building Document Loader, verify the tool is working by retrieving information about the tool:

```bash
pgedge-docloader version
```

Check supported formats:

```bash
pgedge-docloader formats
```

# pgEdge Document Loader

## Overview

The pgEdge Document Loader is a tool written in GoLang, designed to load the
content from a directory or file into the specified columns in a table in a
PostgreSQL database.

## Configuration

The tool will include a -config command line option, to which the user can 
optionally provide a YAML configuration file.

The tool will include command line options, and config file support for all 
configurable functionality, except for the PostgreSQL password which will
be taken (in order of priority) from the PGPASSWORD environment variable, from
the user's .pgpass file (as libpq does), or via interactive command line 
prompt.

The intent of the configuration file is to allow the user to create a reusable
configuration for specific tasks that they expect to repeat. When a 
configuration file is used, all paths (e.g. for the source documents, client
certificates etc.) should be assumed to be relative to the location of the 
configuration file, unless the paths are absolute. If no configuration file is
used, relative paths should be relative to the current working directory.

## Document Support

The user will provide a "-source" option to the tool on the command line or in
the config file. This will be the path either to a single file, a path to a 
directory containing zero or more files, or a glob pattern.

The tool will process either the specified file, or all files in the directory 
or matching the glob pattern, provided they are of a supported type. If an 
unsupported file type is provided as a single file name, the tool will exit 
with a user-friendly error. If an unsupported file type is encountered in a 
directory or from a glob pattern, it will simply be skipped with a 
user-friendly info message to the user.

## Document Formats

The tool will automatically detect the source document format, and convert the
content to Markdown. Where possible, it will extract the document title, e.g.
from the <title> tag in an HTML document, or if the source is Markdown already,
from a line starting with a top-level title marker at the beginning of the 
document (after any metadata).

The tool should support input documents in HTML (.html/.htm), Markdown (.md),
and reStructuredText (.rst).

## Metadata Extraction

The tool will extract (where possible), the filename of the document including
any path (for example, if the -source option is set to docs/ or docs/*.md, the 
filename might be docs/index.md, but if it's set to ./ or ./* it might simply 
be index.md). A command line option (--strip-path) will be provided, in which
case only the actual filename will be recorded.

Additionally, the unaltered source of the document will be extracted, along
with the last modification date (from the filesystem), and of course, the 
converted Markdown text.

## Database Insertion

The user will provide configuration values to connect to the PostgreSQL
database, using either a username and password or client certificates.

They will also provide the name of the database table into which to save the
processed document, along with the names of the columns to use. Where no 
column is provided for a particular piece of data, it will simple be skipped.

Column names for the following pieces of data may be supplied:

* doc_title - Receives the title of the document, where available.
* doc_content - Receives the content of the document.
* source_content - Receives the unmodified source content of the document.
    Must be a bytea column.
* file_name - Receives the name of the source file, including the path if not
    stripped.
* file_created - Receives the timestamp of the source file creation, where 
    available.
* file_modified - Receives the timestamp of the source file's last 
    modification creation, where available.
* row_created - Receives the timestamp of the insertion of the database row.
* row_updated - Receives the timestamp of the last update to the row.

The tool will construct an SQL query for each processed document to perform
the INSERT, inserting into the table and columns specified by the user.

A -update configuration option will also be provided. If this is specified, 
the tool will update a pre-existing row, matched by the filename, if present.
If not present, it will insert a new row.

All database updates/inserts will be performed in a single transaction. If any
processing or database errors occur, the tool will rollback the transaction 
and exit with a useful and user friendly error message.

## Status Summary

When the tool exits, it will print a summary of the work completed.
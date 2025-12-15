# Changelog

All notable changes to pgEdge Document Loader will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0-beta1] - 2025-12-15

### Changed

- Promoted to beta status - all core features complete and tested
- Updated documentation URL to docs.pgedge.com
- Reorganized troubleshooting documentation for better clarity
- Updated navigation tree structure in mkdocs
- Moved license file reference to docs folder

### Fixed

- Fixed config.yml creation example in quickstart guide

## [1.0.0-alpha5] - 2025-12-08

### Changed

- **Documentation restructuring**: Reorganized documentation into separate
  files for better navigation

    - Split supported-formats.md into individual format documentation files
      (html.md, markdown.md, rst.md, sgml.md)
    - Added new documentation pages: authentication.md, best_practices.md,
      metadata.md, quickstart.md, updating.md
    - Renamed unsupported-formats.md to formats.md
    - Updated mkdocs navigation structure

### Fixed

- Fixed repository URL in mkdocs.yml (was pointing to wrong repository)
- Fixed duplicate Reference section in mkdocs navigation
- Added missing trailing newlines to documentation files

## [1.0.0-alpha4] - 2025-12-08

### Changed

- Removed local Claude settings file from version control

## [1.0.0-alpha3] - 2025-12-05

### Added

- **SGML/DocBook support**: New document format support for SGML and DocBook
  XML files (`.sgml`, `.sgm`, `.xml` extensions)

    - Title extraction from `<title>` and `<refentrytitle>` tags
    - DocBook section tags converted to Markdown headings (`<chapter>`,
      `<sect1>`-`<sect5>`, `<refsect1>`-`<refsect3>`, etc.)
    - Code elements converted to inline code (`<literal>`, `<command>`,
      `<filename>`, `<function>`, `<type>`, `<varname>`, `<option>`,
      `<parameter>`, `<constant>`, `<replaceable>`)
    - `<programlisting>` and `<screen>` converted to fenced code blocks
    - `<emphasis>` converted to italic formatting
    - Lists (`<itemizedlist>`, `<orderedlist>`) converted to Markdown lists
    - Links (`<ulink>`) converted to Markdown link format
    - Cross-references (`<xref>`) converted to inline code
    - Special handling for PostgreSQL-style reference pages (`<refentry>`,
      `<refnamediv>`)
    - HTML entities automatically decoded
    - Comments and DOCTYPE declarations stripped

- Changelog documentation

### Changed

- Updated command description to list correct supported formats
- Improved README documentation formatting and consistency

### Fixed

- Command help text incorrectly mentioned PDF support (not implemented)
- Fixed licence URL in documentation

## [1.0.0-alpha2] - 2025-01-20

### Added

- Release workflow using goreleaser for automated builds on release tags

## [1.0.0-alpha1] - 2025-01-15

### Added

- Initial alpha release
- **HTML support**: Convert HTML documents to Markdown

    - Title extraction from `<title>` tag
    - Heading level shifting (h1 → h2, etc.)
    - HTML entity decoding

- **Markdown support**: Pass-through with title extraction

    - Title extraction from first `#` heading
    - YAML frontmatter skipping

- **reStructuredText support**: Convert RST to Markdown

    - Title extraction from underlined headings
    - Heading conversion (both underline and overline+underline styles)
    - Image and figure directive conversion
    - RST directive stripping from titles

- **Database features**:

    - PostgreSQL connection with SSL/TLS support
    - Flexible column mappings
    - Custom metadata columns via `--set-column`
    - Update mode for syncing documents
    - Transactional processing with rollback on failure

- **File processing**:

    - Single file, directory, or glob pattern input
    - Recursive glob matching with `**`
    - Path stripping option
    - Automatic format detection by extension

- **Security**:

    - Password from environment variable (`PGPASSWORD`)
    - Password from `.pgpass` file
    - Interactive password prompt

- **Configuration**:

    - YAML configuration file support
    - Command-line flags for all options

## Next Steps

- [Supported Formats](formats.md) - Full format documentation
- [Configuration](configuration.md) - Configuration options
- [Usage](usage.md) - Usage examples

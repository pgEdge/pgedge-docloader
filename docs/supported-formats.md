# Supported Document Formats

The pgEdge Document Loader supports multiple document formats. All formats
are automatically detected and converted to Markdown.

## HTML Documents

**Extensions:** `.html`, `.htm`

### Processing

- Converted to Markdown using a robust HTML-to-Markdown converter
- Title extracted from `<title>` tag and prepended as `#` heading
- HTML entities automatically decoded (e.g., `&#8212;` becomes `—`)
- All HTML headings shifted down one level:
  - `<h1>` → `##` (level 2)
  - `<h2>` → `###` (level 3)
  - `<h3>` → `####` (level 4)
  - And so on...
- Preserves paragraphs, lists, links, and basic formatting
- Strips scripts, styles, and other non-content elements

### Example

Input HTML:

```html
<!DOCTYPE html>
<html>
<head>
    <title>My Document &#8212; Getting Started</title>
</head>
<body>
    <h1>Introduction</h1>
    <p>This is a <strong>sample</strong> document.</p>
    <h2>Overview</h2>
    <p>More content here.</p>
</body>
</html>
```

Extracted title: `My Document — Getting Started`

Converted Markdown:

```markdown
# My Document — Getting Started

## Introduction

This is a **sample** document.

### Overview

More content here.
```

Note how the title from `<title>` becomes `#`, `<h1>` becomes `##`, and
`<h2>` becomes `###`.

## Markdown Documents

**Extensions:** `.md`

### Processing

- Already in target format - passed through unchanged
- Title extracted from first level-1 heading (`# Title`)
- Skips YAML frontmatter when extracting title
- Preserves all Markdown formatting

### Example

Input Markdown:

```markdown
---
author: John Doe
date: 2024-01-15
---

# My Document

This is the content.
```

Extracted title: `My Document` (YAML frontmatter ignored)

## reStructuredText Documents

**Extensions:** `.rst`

### Processing

- Converted to Markdown format
- Title extracted from underlined headings (first heading found)
- Heading underlines converted to Markdown `#` style
- RST anchors and directives are stripped from both titles and headings
- Both underline-only and overline+underline headings are supported

**Example with directives:**

RST heading with directive:
```rst
`Add Named Restore Point Dialog`:index:
========================================
```

Extracted title: `Add Named Restore Point Dialog` (directive removed)

### Heading Conversion

RST uses underlines (and optionally overlines) to denote headings. The
heading level is determined by the order in which each underline pattern
first appears in the document, not by the specific punctuation character
used.

**Simple underline heading:**

```rst
Main Title
==========

Section
-------
```

**Heading with overline and underline:**

```rst
.. _coding_standards:

*************************
`Coding Standards`:index:
*************************

Sub heading 1
*************
```

Both convert to:

```markdown
# Coding Standards

## Sub heading 1
```

**Key features:**

- Any punctuation character can be used for underlines
- First pattern encountered becomes level 1 (`#`)
- Second pattern becomes level 2 (`##`), and so on
- RST anchors and directives (`.. name:` or `.. _name:`) are automatically
  stripped from the output
- Inline directives (`:index:`, `:ref:`, etc.) are removed from heading
  text
- Headings with overline+underline are distinct from underline-only
  headings

### Image Conversion

RST image and figure directives are converted to Markdown format:

```rst
.. image:: path/to/image.png
   :alt: Image description
   :width: 500px
```

Converts to:

```markdown
![Image description](path/to/image.png)
```

Both `.. image::` and `.. figure::` directives are supported. Alt text is
extracted from the `:alt:` option if present. Other options (width, height,
etc.) are ignored as they are not part of standard Markdown image syntax.

### Limitations

- Complex RST directives are not fully supported
- Only basic conversion is performed
- Advanced features (tables, code blocks with options) may not convert
  perfectly
- Image options like width, height, and alignment are not preserved in the
  Markdown output

## Unsupported Formats

The following formats are **not** supported:

- Microsoft Word (`.doc`, `.docx`)
- OpenDocument (`.odt`)
- Rich Text Format (`.rtf`)
- Plain text (`.txt`)
- LaTeX (`.tex`)

### Handling Unsupported Files

**Single file:** If you specify an unsupported file directly, the tool will
exit with an error.

```bash
$ pgedge-docloader --source file.txt ...
Error: unsupported file type: file.txt
```

**Directory or glob:** Unsupported files are skipped with an informational
message.

```bash
$ pgedge-docloader --source ./docs ...
Processing files from: ./docs
Skipping unsupported file: ./docs/readme.txt
Skipping unsupported file: ./docs/image.png
Processed 10 file(s), skipped 2 file(s)
```

## Checking Supported Formats

List all supported formats:

```bash
$ pgedge-docloader formats
Supported document formats:
  .html
  .htm
  .md
  .rst
```

## Format Detection

Format detection is based solely on file extension (case-insensitive):

- `document.HTML` → Detected as HTML
- `README.MD` → Detected as Markdown
- `guide.RST` → Detected as reStructuredText

Files without extensions or with unknown extensions are treated as
unsupported.

## Best Practices

### HTML Documents

- Ensure documents have a `<title>` tag for proper title extraction
- Use semantic HTML for better Markdown conversion
- Avoid complex layouts that don't translate well to Markdown

### Markdown Documents

- Use a single level-1 heading (`#`) at the top for title extraction
- Place YAML frontmatter before the title if using frontmatter
- Follow standard Markdown syntax for best results

### reStructuredText Documents

- Use standard RST heading underlines
- Avoid complex directives that may not convert well
- Test conversion with sample documents

## Future Format Support

Potential formats for future support:

- Microsoft Word (`.docx`)
- OpenDocument (`.odt`)
- EPUB (`.epub`)
- AsciiDoc (`.adoc`)

## Next Steps

- [Usage](usage.md) - Learn how to process documents
- [Configuration](configuration.md) - Set up column mappings

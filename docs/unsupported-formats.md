# Unsupported Formats

The pgEdge Document Loader supports multiple document formats. The following formats are automatically detected and converted to Markdown.  Format detection is based solely on the file extension (case-insensitive); for details about each supported format type, visit:

- `document.html` → [Detected as HTML](html.md)
- `README.md` → [Detected as Markdown](markdown.md)
- `guide.rst` → [Detected as reStructuredText](rst.md)

!!! note

    Files without extensions or with unknown extensions are treated as unsupported.

The following document formats are **not** supported:

- Microsoft Word (`.doc`, `.docx`)
- OpenDocument (`.odt`)
- Rich Text Format (`.rtf`)
- Plain text (`.txt`)
- LaTeX (`.tex`)

Use the following command to return an up-to-date list of all currently supported formats:

```bash
$ pgedge-docloader formats
Supported document formats:
  .html
  .htm
  .md
  .rst
```

If the Document Loader encounters an unsupported format during a conversion, it handles the request as follows:

**Single file:** If you specify an unsupported file directly, the tool will exit with an error.

```bash
$ pgedge-docloader --source file.txt ...
Error: unsupported file type: file.txt
```

**Directory or glob:** Unsupported files are skipped with an informational message.

```bash
$ pgedge-docloader --source ./docs ...
Processing files from: ./docs
Skipping unsupported file: ./docs/readme.txt
Skipping unsupported file: ./docs/image.png
Processed 10 file(s), skipped 2 file(s)
```

### Future Format Support

Potential formats for future support:

- Microsoft Word (`.docx`)
- OpenDocument (`.odt`)
- EPUB (`.epub`)
- AsciiDoc (`.adoc`)
# Converting and Loading reStructuredText Documents

**Extensions:** `.rst`

During a reStructuredText conversion:

- Converted to Markdown format
- Title extracted from underlined headings (first heading found)
- Heading underlines converted to Markdown `#` style
- RST anchors and directives are stripped from both titles and headings
- Both underline-only and overline+underline headings are supported

!!! note

    - Complex RST directives are not fully supported
    - Only basic conversion is performed
    - Advanced features (tables, code blocks with options) may not convert perfectly
    - Image options like width, height, and alignment are not preserved in the Markdown output


**Example with directives:**

RST heading with directive:
```rst
`Add Named Restore Point Dialog`:index:
========================================
```

Extracted title: `Add Named Restore Point Dialog` (directive removed)

## Heading Conversion

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
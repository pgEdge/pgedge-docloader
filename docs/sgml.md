# Converting and Loading SGML/DocBook Documents

**Extensions:** `.sgml`, `.sgm`, `.xml`

During an SGML conversion:

- documents are converted to Markdown format.
- the title is extracted from `<title>` or `<refentrytitle>` tags (PostgreSQL-style
  reference pages use `<refentrytitle>`).
- DocBook section tags are converted to Markdown headings:

    - `<chapter>`, `<appendix>`, `<article>`, `<book>` → `#` (level 1)
    - `<sect1>`, `<refsect1>`, `<refsynopsisdiv>`, `<section>` → `##`
      (level 2)
    - `<sect2>`, `<refsect2>` → `###` (level 3)
    - `<sect3>`, `<refsect3>` → `####` (level 4)
    - `<sect4>` → `#####` (level 5)
    - `<sect5>` → `######` (level 6)

- Inline code elements are converted to backticks: `<literal>`, `<command>`,
  `<filename>`, `<function>`, `<type>`, `<varname>`, `<option>`,
  `<parameter>`, `<constant>`, `<replaceable>`
- `<programlisting>` and `<screen>` converted to fenced code blocks
- `<emphasis>` converted to italic (`*text*`)
- Lists (`<itemizedlist>`, `<orderedlist>`) are converted to Markdown lists.
- Links (`<ulink>`) are converted to Markdown link format.
- Cross-references (`<xref>`) are converted to inline code with the linkend.
- HTML entities are automatically decoded.
- Comments and DOCTYPE declarations are stripped.

**Example**

Input DocBook:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE book PUBLIC "-//OASIS//DTD DocBook V4.2//EN">
<book>
<title>PostgreSQL Guide</title>
<chapter>
<title>Getting Started</title>
<para>Use the <command>psql</command> command to connect.</para>
<sect1>
<title>Installation</title>
<para>Download from <ulink url="https://postgresql.org">the website</ulink>.</para>
</sect1>
</chapter>
</book>
```

Extracted title: `PostgreSQL Guide`

Converted Markdown:

```markdown
# PostgreSQL Guide

# Getting Started

Use the `psql` command to connect.

## Installation

Download from [the website](https://postgresql.org).
```

### PostgreSQL Reference Pages

The converter includes special handling for PostgreSQL-style reference pages using `<refentry>`:

```xml
<refentry>
<refmeta><refentrytitle>SELECT</refentrytitle></refmeta>
<refnamediv>
<refname>SELECT</refname>
<refpurpose>retrieve rows from a table</refpurpose>
</refnamediv>
<refsect1>
<title>Description</title>
<para>SELECT retrieves rows from tables.</para>
</refsect1>
</refentry>
```

This converts to:

```markdown
# SELECT

## SELECT

retrieve rows from a table

## Description

SELECT retrieves rows from tables.
```

### Limitations

- Not all DocBook elements are fully supported
- Complex nested structures may not convert perfectly
- Only basic conversion is performed for most elements
- Tables and complex formatting may require manual adjustment

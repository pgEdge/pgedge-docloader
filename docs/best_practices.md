# Best Practices

When preparing documents for extraction, you should ensure that each document type has the expected properties.

**HTML Documents**

HTML documents should:

- have a `<title>` tag for proper title extraction.
- use semantic HTML for efficient Markdown conversion.
- avoid complex layouts that don't translate well to Markdown.

**Markdown Documents**

Markdown documents should:

- use a single level-1 heading (`#`) at the top of each file for title extraction.
- place YAML frontmatter before the title (if using frontmatter).
- follow standard Markdown syntax for best results.

**reStructuredText Documents**

reStructuredText documents should:

- use standard RST heading underline formats.
- avoid complex directives that may not convert well.
- test conversion with sample documents.


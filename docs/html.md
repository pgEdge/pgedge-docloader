# Converting and Loading HTML Documents

**Extensions:** `.html`, `.htm`

During an HTML conversion:

- documents are converted to Markdown using a robust HTML-to-Markdown converter
- the title is extracted from `<title>` tag and prepended as `#` heading
- HTML entities automatically decoded (e.g., `&#8212;` becomes `—`)
- All HTML headings shifted down one level:
  - `<h1>` → `##` (level 2)
  - `<h2>` → `###` (level 3)
  - `<h3>` → `####` (level 4)
  - And so on...
- Preserves paragraphs, lists, links, and basic formatting
- Strips scripts, styles, and other non-content elements

**Example**

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

Note how the title from `<title>` becomes `#`, `<h1>` becomes `##`, and `<h2>` becomes `###`.
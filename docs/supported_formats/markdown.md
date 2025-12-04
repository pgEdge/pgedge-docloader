# Loading Markdown Documents

**Extensions:** `.md`

During a Markdown conversion:

- Files are already in target format - passed through unchanged
- Title extracted from first level-1 heading (`# Title`)
- Skips YAML frontmatter when extracting title
- Preserves all Markdown formatting

**Example**

Input Markdown:

```markdown
---
author: John Doe
date: 2024-01-15
---

# My Document

This is the content.
```

The extracted title is: `My Document` (YAML frontmatter is ignored).
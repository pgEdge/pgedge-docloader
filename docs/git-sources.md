# Using Git Repository Sources

As an alternative to local files, pgEdge Document Loader can clone and process
documentation directly from Git repositories. This is useful for:

- Loading documentation from remote repositories without manual cloning
- Processing specific branches or tags (e.g., versioned documentation)
- Automated pipelines that fetch and load docs from source control

## Git Source Options

| Option           | Required | Description                                      |
|------------------|----------|--------------------------------------------------|
| `--git-url`      | Yes*     | Git repository URL to clone                      |
| `--git-branch`   | No       | Branch to checkout (default: repository default) |
| `--git-tag`      | No       | Tag to checkout (mutually exclusive with branch) |
| `--git-doc-path` | No       | Path within repository to process (repeatable)   |
| `--git-clone-dir`| No       | Directory to store cloned repositories           |
| `--git-keep-clone`| No      | Keep cloned repository after processing          |
| `--git-skip-fetch`| No      | Skip fetch if repository already exists          |

*Either `--source` or `--git-url` is required, but not both.

## Basic Usage

Clone a repository and process all supported files from the root:

```bash
pgedge-docloader \
    --git-url https://github.com/org/docs-repo.git \
    --db-host localhost \
    --db-name mydb \
    --db-user myuser \
    --db-table documents \
    --col-doc-content content \
    --col-file-name filename
```

## Processing a Specific Directory

Use `--git-doc-path` to process files from a specific directory within the
repository:

```bash
pgedge-docloader \
    --git-url https://github.com/org/project.git \
    --git-doc-path docs/api \
    --db-host localhost \
    --db-name mydb \
    --db-user myuser \
    --db-table documents \
    --col-doc-content content
```

The `--git-doc-path` option supports glob patterns:

```bash
# Process only markdown files in the docs directory
pgedge-docloader \
    --git-url https://github.com/org/project.git \
    --git-doc-path "docs/**/*.md" \
    --config config.yml
```

## Multiple Source Patterns

You can specify multiple `--git-doc-path` options to process files from
different locations:

```bash
# Process both docs directory and root-level markdown files
pgedge-docloader \
    --git-url https://github.com/org/project.git \
    --git-doc-path "docs/**/*.md" \
    --git-doc-path "*.md" \
    --config config.yml
```

In a configuration file, use a YAML list:

```yaml
git-url: https://github.com/org/project.git
git-doc-path:
    - "docs/**/*.md"
    - "*.md"
```

## Working with Branches and Tags

### Checkout a Specific Branch

```bash
pgedge-docloader \
    --git-url https://github.com/org/docs.git \
    --git-branch main \
    --git-doc-path docs \
    --config config.yml
```

### Checkout a Specific Tag

Use tags for versioned documentation:

```bash
pgedge-docloader \
    --git-url https://github.com/org/project.git \
    --git-tag v2.0.0 \
    --git-doc-path docs \
    --set-column version="2.0.0" \
    --config config.yml
```

!!! note

    `--git-branch` and `--git-tag` are mutually exclusive. You cannot specify
    both options at the same time.

## Persistent Clone Directory

By default, repositories are cloned to a temporary directory and removed after
processing. For repeated runs, you can specify a persistent clone directory:

```bash
pgedge-docloader \
    --git-url https://github.com/org/docs.git \
    --git-clone-dir /var/cache/docloader/repos \
    --git-keep-clone \
    --config config.yml
```

On subsequent runs with `--git-skip-fetch`, the tool will reuse the existing
clone without fetching updates:

```bash
pgedge-docloader \
    --git-url https://github.com/org/docs.git \
    --git-clone-dir /var/cache/docloader/repos \
    --git-keep-clone \
    --git-skip-fetch \
    --config config.yml
```

## Configuration File Example

Git source options can also be specified in a configuration file:

```yaml
# Git source configuration
git-url: https://github.com/org/docs-repo.git
git-branch: main
git-doc-path:
    - "docs/**/*.md"
    - "*.md"
git-clone-dir: /var/cache/docloader/repos
git-keep-clone: true

# Database configuration
db-host: localhost
db-name: mydb
db-user: myuser
db-table: documents

# Column mappings
col-doc-content: content
col-file-name: filename
col-doc-title: title

# Custom metadata
custom-columns:
    source: "git-repo"
    project: "my-project"
```

Then run with:

```bash
pgedge-docloader --config config.yml
```

## Authentication

### HTTPS URLs

For public repositories, use the HTTPS URL directly:

```bash
--git-url https://github.com/org/public-repo.git
```

For private repositories, you can use a personal access token in the URL:

```bash
--git-url https://TOKEN@github.com/org/private-repo.git
```

Or configure Git credential helpers before running the tool.

### SSH URLs

For SSH authentication, ensure your SSH keys are configured:

```bash
--git-url git@github.com:org/repo.git
```

## Error Handling

The tool will fail with a clear error message if:

- Git is not installed on the system
- The repository URL is invalid or inaccessible
- The specified branch or tag does not exist
- The `--git-doc-path` does not exist in the repository

## Best Practices

1. **Use tags for versioned docs**: When loading documentation for specific
   software versions, use `--git-tag` to ensure consistency.

2. **Cache clones for repeated runs**: Use `--git-clone-dir` and
   `--git-keep-clone` to avoid re-cloning on every run.

3. **Use `--git-skip-fetch` carefully**: Only skip fetching when you're sure
   the local clone is up-to-date.

4. **Set version metadata**: Use `--set-column` to add version information
   when processing tagged releases:

    ```bash
    --git-tag v1.2.3 --set-column version="1.2.3"
    ```

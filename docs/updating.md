# Updating a Stored Document

When invoked with the `--update` flag, the tool reviews previously stored documents and updates existing rows (matched by filename) or inserts new ones:

```bash
pgedge-docloader \
  --config config.yml \
  --update
```
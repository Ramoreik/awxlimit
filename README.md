# Awx Limit

Resolve Ansible-style `--limit` / Tower Job Template *Limit* patterns to concrete hostnames.

## Inventory format (JSON)

```json
{
  "hosts": ["web01", "web02", "db01", "phoenix"],
  "groups": [
    {"name": "webservers", "hosts": ["web01", "web02"]},
    {"name": "dbservers", "hosts": ["db01", "phoenix"]},
    {"name": "staging", "hosts": ["web02", "db01", "phoenix"]},
    {"name": "all_apps", "hosts": ["web01"], "children": ["webservers", "dbservers"]}
  ]
}
```

## Semantics

- Union (any): `a:b` or `a,b`
- Intersection (all): `&c`
- Exclude (not): `!d`

If there are **no union** tokens, the base set is **all inventory hosts**, and then `&` / `!` are applied.

## Run

```bash
cat > inv.json <<'JSON'
{
  "hosts": ["web01", "web02", "db01", "phoenix", "misc01"],
  "groups": [
    {"name": "webservers", "hosts": ["web01", "web02"]},
    {"name": "dbservers", "hosts": ["db01", "phoenix"]},
    {"name": "staging", "hosts": ["web02", "db01", "phoenix"]}
  ]
}
JSON

go run ./cmd/awxlimit -pattern "webservers:dbservers:&staging:!phoenix" -inventory inv.json
```

## Notes

- Tokens are resolved by:
  - exact group name => group expansion (including child groups)
  - exact hostname
  - glob patterns (`*`, `?`, `[]`) applied to hosts and group names
  - regex patterns (`~...`) applied to hosts and group names
- This is intentionally a **minimal** inventory model; it does not cover every Ansible feature.

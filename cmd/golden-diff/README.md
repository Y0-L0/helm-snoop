# golden-diff

Compares `.golden.json` files in the working tree against their versions at a git commit. Useful for reviewing how changes to helm-snoop affect analysis results.

## Usage

```bash
go run ./cmd/golden-diff [flags] [file.golden.json ...]
```

With no files specified, it automatically finds and diffs all modified `.golden.json` files.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-commit` | `HEAD` | Git ref to compare against |

## Examples

```bash
# Diff all changed golden files against HEAD
go run ./cmd/golden-diff

# Diff against a specific commit or branch
go run ./cmd/golden-diff -commit 90f67ab
go run ./cmd/golden-diff -commit main

# Diff specific files only
go run ./cmd/golden-diff testdata/foo.golden.json testdata/bar.golden.json
```

## Output

For each file, the tool compares three path categories (Referenced, Unused, Undefined) and reports:

```
=== Unused Paths ===

Only in [1] (2 paths):              ← in the commit but removed from working tree
  - .Values.foo (kinds: map)
  - .Values.bar (kinds: scalar)

Only in [2] (1 paths):              ← new in working tree
  + .Values.baz (kinds: scalar)

Paths with different kinds (1):
  ~ .Values.qux
      [1]: scalar
      [2]: map,scalar

Common paths: 42
```

Golden files not found in the comparison commit are treated as empty (all paths show as added).

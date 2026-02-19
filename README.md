# helm-snoop

helm-snoop keeps your config and docs in your values.yaml
and (in the future) schema.json
in sync with the code in your helm template files. \
It parses the values.yaml and template files of helm charts
and finds undefined and unused values.

**Beta Status:** Expect some false positives and breaking changes
as the project matures.

## 🚀 Getting Started

### 🐳 Try it out

```bash
docker run --rm -v $(pwd):/chart ghcr.io/y0-l0/helm-snoop:latest /chart
```

### 📦 Installation

**Homebrew (macOS/Linux):**
```bash
brew install y0-l0/tap/helm-snoop
```

**Binary (Linux amd64):**
```bash
VERSION=0.2.0
curl -fsSL "https://github.com/y0-l0/helm-snoop/releases/download/v${VERSION}/helm-snoop_${VERSION}_linux_amd64.tar.gz" | tar -xz
sudo mv helm-snoop /usr/local/bin/
```

## 💻 Usage

```bash
# Minimal
helm-snoop <path-to-chart>

# With all optional flags
helm-snoop --ignore /image/tag --ignore /config/* --json --referenced -vv <path-to-chart>
```

Analyzes Helm charts and reports:
- **Unused:** Keys in values.yaml never used in templates
- **Undefined:** Paths used in templates but not defined in values.yaml

See [docs/CLI.md](docs/CLI.md) for complete documentation.

### 🪝 pre-commit / prek

Add to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/y0-l0/helm-snoop
    rev: v0.2.0
    hooks:
      - id: helm-snoop        # requires helm-snoop on PATH
      # - id: helm-snoop-docker  # uses Docker instead
```

Two hooks are available: `helm-snoop` (requires the binary on `$PATH`) and
`helm-snoop-docker` (runs via `ghcr.io/y0-l0/helm-snoop`). Changed files are
resolved to their parent chart directories automatically.

### ✅ Features

- **Variable tracking:** Variables are tracked across references (e.g., `{{ $var := .Values.foo }}{{ $var.bar }}`)
- **Context-aware path resolution:** Correctly resolves relative paths within `with` and `range` contexts (e.g., `.Values.config` → `with` → `.timeout` resolves to `.Values.config.timeout`)
- **Dict/list operations:** Tracks values through `dict`, `list`, `merge`, `concat` operations
- **Nested template definitions:** Follows `define` blocks across multiple files
- **Include/Template functions:** Template includes are followed and analyzed
- **Wildcard ignore patterns:** Advanced pattern matching for suppressing specific warnings
- **Control flow:** All branches of `if/else` blocks are analyzed

### ⚠️ Limitations

- **Limited `tpl` function support:** Dynamic template strings have partial support
- **No schema.json validation:** Only compares templates against values.yaml, not against schema definitions
- **Limited dynamic evaluation:** Complex patterns like `{{ index .Values.a .Values.b }}` may not be fully resolved
- **No subchart analysis:** Does not analyze subcharts (only collects template functions via `define`)
- **No global values from subcharts:** May report false positives for undefined global paths from subcharts

## 🔍 Background

### 🎯 The Problem

Golang doesn't compile with undefined or unused variables. Helm has no such checks.

In my day job, my team maintains 50_000 lines of helm chart code.

The majority of bugs in these helm charts are:
- Typos in either template files or values.yaml files
- Mismatches between defaults/documentation in values.yaml and actual template implementation

It is also close to impossible to keep the documentation in values.yaml
in sync with the actual template implementation.
New config options are implemented but never documented.
Old config options are never removed from the documentation.

All of these could be detected by a Helm Chart linter using static analysis:
- No undefined values - `.Values.*` paths require defaults and documentation in values.yaml
- No unused values - No outdated or misspelled keys in values.yaml

helm-snoop is (trying to become) that linter.

### 🔗 Related Tools

**helm-unused-values** Similar PoC implementation, appears unfinished and inactive since some time.

**Regex-based tools** Limited by their approach; can't handle complex template logic.

**JSON Schema for values.yaml** Validates user input, which is valuable. However, it suffers from the same sync problem: if your schema.json is out of sync with your template implementation, users can't configure those out-of-sync properties.

For detailed comparison, see [docs/RESEARCH.md](docs/RESEARCH.md)

## 🤝 Contributing

Contributions would be wonderful:
- Architecture input, especially from Helm codebase experts
- Testing and user feedback
- Code quality and idiomatic Go (inexperienced gopher here!)

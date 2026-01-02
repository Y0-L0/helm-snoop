# helm-snoop

helm-snoop keeps your config and docs in your values.yaml
and (in the future) schema.json
in sync with the code in your helm template files. \
It parses the values.yaml and template files of helm charts
and finds undeclared and unused values.

**Alpha Status:** Expect false positives, false negatives, and breaking changes
as the project matures.

Contributions would be wonderful:
- Architecture input, especially from Helm codebase experts
- Testing and user feedback
- Code quality and idiomatic Go (inexperienced gopher here!)
- Code for the CLI or `tpl` implementation

## 🚀 Usage

```bash
helm-snoop <path-to-chart> [log-level]
```

The chart-path can be a directory or .tgz file. \
Log-level options: debug|info|warn|error

Analyzes a Helm chart and reports:
- Referenced - .Values paths defined in values.yaml and used in templates
- Defined-not-used - Unused: Keys in values.yaml that templates never use
- Used-not-defined - Undefined: Paths that templates use but values.yaml doesn't define

## ⚠️ Current Limitations

- No `tpl` function support. Dynamic template strings aren't evaluated.
- Variables aren't tracked across references. e.g. `{{ $var := .Values.foo }}{{ $var.bar }}` doesn't resolve to `.Values.foo.bar`
- No schema.json validation. Only compares templates against values.yaml, not against schema definitions.
- No support for dynamic evaluation. e.g. `{{ index .Values.a .Values.b }}`
- No wildcard matching for functions like `toYaml` or `range`. e.g. `{{ toYaml .Values.a }}` currently does not match the defined `a.b.c`. -> False positive.
- Analyze subcharts. (not just collect template functions (`define`))
- Respect / Include globals from subcharts. -> false positive undefined paths.

## 🔍 The Problem

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

## 🔗 Related Tools

**helm-unused-values** Similar PoC implementation, appears unfinished and inactive since some time.

**Regex-based tools** Limited by their approach; can't handle complex template logic.

**JSON Schema for values.yaml** Validates user input, which is valuable. However, it suffers from the same sync problem: if your schema.json is out of sync with your template implementation, users can't configure those out-of-sync properties.

For detailed comparison, see [docs/RESEARCH.md](docs/RESEARCH.md)

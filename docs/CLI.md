# CLI Documentation

## Usage

```bash
helm-snoop [FLAGS] <chart-path>
helm-snoop version
```

**Arguments:**
- `<chart-path>`: Path to Helm chart directory (required, e.g., `.` or `./my-chart`)

## Subcommands

### `version`
Display version information.

```bash
helm-snoop version
```

## Flags

### `-i, --ignore <PATTERN>`
Ignore findings for specific path patterns. Supports wildcards and advanced matching. Suppresses both output and exit code impact. Can be specified multiple times.

```bash
helm-snoop -i /image/tag -i /config/* ./my-chart
```

**Pattern Syntax:**
- Must start with `/` (JSON Pointer style)
- Supports wildcards with `*`
- Integer segments match both array indices and string keys
- Examples:
  - `/image/tag` - Exact match
  - `/config/*` - All descendants under config
  - `/items/0` - Matches both `items[0]` and `items["0"]`
  - `/a/*/c` - Interior wildcard (one level)

See [ignore-patterns.md](ignore-patterns.md) for complete pattern documentation.

### `-s, --strict`
Enable strict mode for enhanced static analysis validation. **(NOT YET IMPLEMENTED)**

Planned behavior: Enforce complete static resolvability, fail on dynamic key access patterns.

### `-e, --exclude-subchart <CHART-NAME>`
Exclude specific subcharts from analysis. Can be specified multiple times. **(NOT YET IMPLEMENTED)**

```bash
helm-snoop -e "postgresql" -e "redis" ./my-chart
```

### `-r, --recursive`
Analyze subcharts recursively. Without this flag, only the root chart is analyzed (template definitions are always followed regardless). **(NOT YET IMPLEMENTED)**

```bash
helm-snoop -r ./my-chart
```

### `--json`
Output results in JSON format for machine-readable consumption (CI/CD integration).

```bash
helm-snoop --json ./my-chart
```

### `--referenced`
Include referenced values in the output. By default, only findings (defined-not-used and used-not-defined) are shown.

```bash
helm-snoop --referenced ./my-chart
```

### `-f, --values <PATH>`
Specify path to values file (overrides default `values.yaml` in chart directory). **(NOT YET IMPLEMENTED)**

```bash
helm-snoop -f ./custom-values.yaml ./my-chart
```

### `-v`
Increase log/verbosity level. Can be repeated for more verbose output.

- No `-v`: Warnings only (default)
- `-v`: Info level
- `-vv` or more: Debug level

```bash
helm-snoop ./my-chart            # Default: warnings
helm-snoop -v ./my-chart         # Info level
helm-snoop -vv ./my-chart        # Debug level
helm-snoop -vvv ./my-chart       # Debug level
```

## Exit Codes

- `0`: Success, no findings
- `1`: Issues found (unused or undefined values) or analysis error

**Note:** Ignored findings (via `-i`) do not affect exit code.

## Implementation

Uses `cobra` for CLI parsing with the `version` subcommand.

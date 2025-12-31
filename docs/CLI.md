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

### `-i, --ignore <KEY>`
Ignore findings for specific keys (exact match). Suppresses both output and exit code impact. Can be specified multiple times.

```bash
helm-snoop -i "image.tag" -i "legacy.setting" ./my-chart
```

### `-s, --strict`
Enable strict mode for enhanced static analysis validation. **(Reserved for future implementation)**

Planned behavior: Enforce complete static resolvability, fail on dynamic key access patterns.

### `-e, --exclude-subchart <CHART-NAME>`
Exclude specific subcharts from analysis. Can be specified multiple times.

```bash
helm-snoop -e "postgresql" -e "redis" ./my-chart
```

### `-r, --recursive`
Analyze subcharts recursively. Without this flag, only the root chart is analyzed (template definitions are always followed regardless).

```bash
helm-snoop -r ./my-chart
```

### `--json`
Output results in JSON format for machine-readable consumption (CI/CD integration).

```bash
helm-snoop --json ./my-chart
```

### `-f, --values <PATH>`
Specify path to values file (overrides default `values.yaml` in chart directory).

```bash
helm-snoop -f ./custom-values.yaml ./my-chart
```

### `-v`
Increase log/verbosity level. Can be repeated for more verbose output.

- No `-v`: Errors and warnings (default)
- `-v`: Errors, warnings, and info
- `-vv`: Errors, warnings, info, and debug
- `-vvv`: Errors, warnings, info, debug, and trace

```bash
helm-snoop ./my-chart            # Default: errors and warnings
helm-snoop -v ./my-chart         # Add info level
helm-snoop -vv ./my-chart        # Add debug level
helm-snoop -vvv ./my-chart       # Add trace level
```

## Exit Codes

- `0`: Success, no findings
- `1`: Issues found (unused or undefined values)
- `2`: Analysis error or invalid usage

**Note:** Ignored findings (via `-i`) do not affect exit code.

## Implementation

Can use either Go stdlib `flag` or a third-party library (e.g., `cobra`, `cli`). No subcommands required.

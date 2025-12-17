# Helm Unused Keys Analyzer — Implementation Guide

This document defines an MVP-first path to validate the approach, followed by a feature wishlist for the complete implementation. Numbered lists are used throughout as requested.

## Two-Stage MVP Plan

1. CLI stub
   1. Accept a single positional argument: the Helm chart directory path.
   2. Print results to stdout; no flags, JSON, or exit-code policies yet.

2. Stage 1: Minimal static extraction
   1. Load the chart and parse a single simple template (for example, a ConfigMap) that uses direct `.Values` field chains only.
   2. Walk the AST and extract `.Values.foo.bar` field chains from `ActionNode`/`NodeChain`.
   3. Flatten `values.yaml` and compute:
      1. Referenced keys (direct only).
      2. Defined keys.
      3. Report “used-not-defined” and “defined-not-used” as plain text.

3. Stage 2: Static functions and helpers (still static)
   1. Extend extraction to handle simple function forms that wrap value reads without changing the accessed path:
      1. `quote`, `upper`, `lower` (string wrappers).
      2. `default X Y` (treat as a read of `Y`).
   2. Add include/template graph for static helpers:
      1. Parse `_helpers.tpl` and collect `define` blocks.
      2. Resolve `include`/`template` calls and union used keys across files.
   3. Validate on a helper chain such as `common.names.fullname` to ensure call graph resolution works without dynamic keys.

4. MVP fixtures and checks
   1. Fixture A (ConfigMap): direct `.Values` only; verify referenced vs defined sets.
   2. Fixture B (Functions): add `quote`/`default` around direct paths; results unchanged for used keys.
   3. Fixture C (Helpers): use `include` with static helper templates; verify union of used keys across files.

5. MVP reporting
   1. Text-only output with three sections: Referenced, Defined-not-used, Used-not-defined.
   2. Include filename:line when available for referenced sites.

## Alternative Validation Approaches

1. Cross-check against helm-values-check
   1. Run both tools on the same fixtures and compare sets; differences reveal gaps quickly.

2. Golden snapshots from real charts
   1. Capture `helm-snoop` output on a few public charts; review changes as functionality grows.

3. Property-based snippet tests
   1. Generate tiny templates combining direct paths with `quote/default/include`; ensure the referenced set is stable under benign wrappers.

4. Schema-on-demand (later)
   1. For charts without schemas, generate a minimal schema from defaults and verify that MVP output remains stable when a schema is introduced.

## Feature Wishlist (Post-MVP)

1. Template parsing and graph
   1. Parse all templates under `templates/**` (and `_helpers.tpl`) into Go `text/template/parse` ASTs.
   2. Collect all `define` blocks by name and track file origins for diagnostics.
   3. Detect `template`/`include` calls; build a call graph with recursion guard.

2. Core extractor (comprehensive)
   1. Walk ASTs visiting all branches (`IfNode`, `WithNode`, `RangeNode` → visit `List` and `ElseList`).
   2. Extract direct field chains `.Values.foo.bar` from `ActionNode` pipelines and `NodeChain`.
   3. Record referenced Paths; flatten to dot notation for reporting.

3. Variables and scope
   1. Implement a scope stack for variables (`:=` declarations and `=` assignments`).
   2. Support aliasing `$x := .Values.a.b` and chained usage `$x.c`.
   3. Respect scoping rules in `if/with/range/define` bodies.

4. Functions (literals)
   1. Implement evaluators for `index`/`dig`/`get` when all keys are string literals.
   2. Implement `default`/`required`: treat as reads of the non-default operand.
   3. Record unresolved cases as dynamic sites for later reporting.

5. Schema integration (lookup)
   1. Load and parse `values.schema.json` when present.
   2. Provide helpers to resolve path existence/type, object closedness (`additionalProperties: false`), and string enums.
   3. Use enums to resolve variables to `LiteralSet` where applicable.

6. Functions (schema-aware)
   1. Extend `index`/`dig` evaluators to accept key variables if they resolve to finite `LiteralSet` via schema (enums or closed objects).
   2. Range handling: mark containers used for `range` over objects/arrays; in `--strict` and closed objects, mark enumerated children as used.

7. tpl handling
   1. Analyze `tpl` only when the argument is literal or sourced from default `values.yaml`; parse and analyze recursively.
   2. When `tpl` source could be user override or cannot be proven literal: warn (error in `--strict`).

8. Reporting and CLI
   1. Compute sets: `unused` (defined-only), `undefined` (referenced-only), and `dynamic` (unresolved).
   2. Render text and JSON reports with locations (file:line) and implement exit codes based on fail-on categories.
   3. Add flags: `--values`, `--schema`, `--strict`, `--format`, `--include-notes`, `--fail-on`.

9. Subcharts and options
   1. Optionally traverse `charts/**` for subcharts; analyze their templates and values.
   2. Merge or separate reports per chart; add flags as needed.
   3. Support `--include-notes` to include/exclude `NOTES.txt`.

10. Testing strategy (expanded)
   1. Unit tests: pipeline evaluators (`index`, `dig`, `get`, `default`), variable aliasing, field chains.
   2. Schema helpers: path resolution, enum extraction, closed object detection.
   3. AST traversal tests: visit both branch bodies and follow include/template with recursion guard.
   4. End-to-end tests: run `helm-snoop` on fixture charts; snapshot text/JSON outputs; verify `--strict` failures for unresolved dynamic sites.

## Developer Notes

1. Memoize schema path lookups; cache parsed templates to keep runs fast.
2. Normalize paths to dot notation consistently; differentiate container vs leaf usage in reports.
3. Keep evaluators conservative: prefer `Unknown` over incorrect precision; rely on `--strict` to enforce completeness.

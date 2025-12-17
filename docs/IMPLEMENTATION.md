# Helm Unused Keys Analyzer — Implementation Guide

This document describes the phased implementation plan, testing approach, and practical guidance to build the analyzer according to the architecture.

## Phased Plan

1) Bootstrap
- Create CLI skeleton and project layout.
- Load chart directory; read `values.yaml` and optional `values.schema.json`.
- Wire basic flags: `--values`, `--schema`, `--strict`, `--format`, `--include-notes`, `--fail-on`.

2) Template Parsing & Graph
- Parse all templates under `templates/**` (and `_helpers.tpl`) into Go `text/template/parse` ASTs.
- Collect all `define` blocks by name; track file origins for diagnostics.
- Detect `template`/`include` calls; prepare a simple call graph and recursion guard.

3) Core Extractor (Static Basics)
- Walk ASTs visiting all branches (`IfNode`, `WithNode`, `RangeNode` → visit `List` and `ElseList`).
- Extract direct field chains `.Values.foo.bar` from `ActionNode` pipelines.
- Record referenced Paths; flatten to dot notation for reporting.

4) Variables & Scope
- Implement a scope stack for variables (`:=` declarations and `=` assignments).
- Support aliasing `$x := .Values.a.b` and chained usage `$x.c`.
- Respect scoping rules in `if/with/range/define` bodies.

5) Functions (Literals)
- Implement evaluators for `index`/`dig`/`get` when all keys are string literals.
- Implement `default`/`required`: treat as reads of the non-default operand.
- Record unresolved cases as dynamic sites for later reporting.

6) Schema Integration (Lookup)
- Load and parse `values.schema.json` when present.
- Provide helpers to resolve path existence/type, object closedness (`additionalProperties: false`), and string enums.
- Use enums to resolve variables to `LiteralSet` where applicable.

7) Functions (Schema-Aware)
- Extend `index`/`dig` evaluators to accept key variables if they resolve to finite `LiteralSet` via schema (enums or closed objects).
- Range handling:
  - Mark containers used for `range` over objects/arrays.
  - In `--strict` and closed objects: mark enumerated children as used.

8) tpl Handling
- Allow analyzing `tpl` only when the argument is literal or sourced from default `values.yaml`; parse and analyze recursively.
- When `tpl` source could be user override or cannot be proven literal: warn (error in `--strict`).

9) Reporting & CLI Wiring
- Compute sets: `unused` (defined-only), `undefined` (referenced-only), and `dynamic` (unresolved).
- Render text and JSON reports with locations (file:line) where possible.
- Implement exit codes based on `--fail-on` categories.

10) Subcharts & Options
- Optionally traverse `charts/**` for subcharts; analyze their templates and values.
- Merge or separate reports per chart; add flags as needed.
- Support `--include-notes` to include/exclude `NOTES.txt`.

11) Tests & Fixtures
- Create chart fixtures to cover:
  - Direct paths; variable aliasing.
  - `index`/`dig` with literals and enum-driven keys.
  - Ranges over arrays/objects (closed vs open schema).
  - `include`/`define` graphs; recursion guard.
  - `tpl` literal/default-only vs override-origin.
  - Subcharts.
- Golden tests for text/JSON outputs and exit codes.

## Testing Strategy
- Unit tests
  - Pipeline evaluators for `index`, `dig`, `get`, `default`, variable aliasing, and field chains.
  - Schema helpers (path resolution, enum extraction, closed object detection).
- AST traversal tests
  - Ensure both branch bodies are visited and scoped correctly.
  - Ensure template/include references are followed with recursion guard.
- End-to-end tests
  - Run analyzer on fixture charts; snapshot text/JSON outputs.
  - Verify `--strict` mode failures for unresolved dynamic sites.

## Developer Notes
- Memoize schema path lookups; cache parsed templates to keep runs fast.
- Normalize paths to dot notation consistently; differentiate container vs leaf usage in reports.
- Keep evaluators conservative: prefer `Unknown` over incorrect precision; rely on `--strict` to enforce completeness.

## Milestones
- M1: Basic extractor (direct paths) + text report.
- M2: Variables + literal `index/dig/get` + JSON output.
- M3: Schema integration (enums/closed objects) + strict mode basics.
- M4: Range handling + include/template call graph.
- M5: `tpl` handling + subcharts + CI-ready exit codes.


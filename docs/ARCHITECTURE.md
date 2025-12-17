# Helm Unused Keys Analyzer — Architecture & Implementation Plan

## Overview
A static analysis tool that detects drift between Helm chart templates and values by:
- Extracting all referenced `.Values` keys from Go templates (including all branches).
- Flattening keys defined in `values.yaml` and optionally `values.schema.json`.
- Reporting differences: unused keys (defined but never referenced) and undefined keys (referenced but not defined by defaults/schema).

`helm-snoop` is static-first, schema-aware, and offers a `--strict` mode to enforce full static resolvability in common dynamic patterns.

## Scope
- Charts: root and subcharts (optional; see Implementation Plan phase for order).
- Files: `templates/**`, `_helpers.tpl`, `values.yaml`, `values.schema.json` (if present).
- Outputs: machine-readable and human-readable reports with exit codes suitable for CI.

## Non-Goals (initial)
- Proving semantic reachability (detecting dead template code).
- Fully interpreting arbitrary dynamic value access (we flag/warn, or require schema rigidity under `--strict`).
- Executing templates or connecting to Kubernetes APIs.

## High-Level Approach
1. Parse all templates into the Go template AST (`text/template/parse`).
2. Build a call graph for `define`/`template` and track `include` usage.
3. Walk the AST for every template, visiting all branches (`if`, `with`, `range`).
4. Evaluate pipelines to extract `.Values` references, including function forms like `index` and `dig` when keys are statically known.
5. Track variable assignments and scope to resolve common aliasing.
6. Integrate schema knowledge to resolve dynamic keys when possible (e.g., enums, closed object properties).
7. Produce sets:
   - Referenced keys (precise where possible; ranges over objects/lists mark container usage; dynamic unresolved marked separately).
   - Defined keys (from `values.yaml` and optionally `values.schema.json`).
8. Report unused and undefined keys; warn on unresolved dynamic accesses.

## Parsing & AST
- Prefer Helm chart loader: use `helm.sh/helm/v3/pkg/chart/loader` to load charts, values, templates, and subcharts consistently.
- Use Go `text/template` parser to obtain `parse.Tree` per file.
- Node types of interest: `ActionNode` (pipelines), `IfNode`, `RangeNode`, `WithNode`, `TemplateNode`, `ListNode`.
- Visit both `List` and `ElseList` for branching nodes; branch reachability is ignored by design.
- Collect `define` blocks; map them by name for resolving `template`/`include` references.
- Guard against recursive includes/templates with a recursion limit; surface a diagnostic when cycles are detected.

## Function Evaluators
- Maintain a registry of lightweight evaluators keyed by function name (e.g., `index`, `dig`, `get`, `default`, `hasKey`, `include`, `tpl`).
- Keep the walker generic; evaluators interpret pipelines into the abstract domain (Path/PathSet/LiteralSet/Unknown).
- Provide a tolerant FuncMap (Sprig + stubs) for parsing only; do not execute functions during analysis.

## Data Model (Abstract Values)
Represent expressions in pipelines using a small abstract domain:
- Path: root + path segments, e.g., `Root=Values, ["image","repository"]`.
- PathSet: finite set of Paths (for union results, e.g., enum-driven keys).
- LiteralSet: finite set of strings (used for dynamic key operands).
- Unknown: cannot be resolved statically to a finite set.
- Container markers: Object(root path), Array(root path) for ranges over maps/lists.

Utilities:
- FlattenedKey: dot-form string for Paths (e.g., `image.repository`).
- Scope: symbol table mapping var names → (Path | PathSet | LiteralSet | Unknown).

## Analysis Rules (selected)
- Field chain: `.Values.a.b` → Path(Values, [a,b]).
- Variable alias: `$x := .Values.a.b` → bind `$x` to that Path.
- Variable literal: `$k := "foo"` → LiteralSet{"foo"}.
- Enum value: `$k := .Values.someEnum` with schema enum → LiteralSet{...}.
- index:
  - `index <root> <k1> <k2> ...` → if all keys resolve to `LiteralSet`, produce `PathSet` of all combinations; else Unknown.
  - `<root>` may be `.Values` or a `Path`/`PathSet`.
- dig (Sprig):
  - `dig "a" "b" .Values` or `.Values | dig "a" "b"` → same as index with all-literal keys.
- default:
  - `default X Y` where `Y` resolves to a Path/PathSet → treat as a read of `Y`.
- get/hasKey/keys (Sprig):
  - `get .Values "a"` reads `.Values.a` if literal; `hasKey` reads container `.Values`/`.Values.a` (optional: mark as guarded read).
- range over object:
  - `range $k, $v := .Values.annotations` → mark `.Values.annotations` container as used. In strict+closed schema: count enumerated children as used.
- range over array:
  - `range .Values.imagePullSecrets` → mark `.Values.imagePullSecrets` container as used.
- include/template:
  - `include "name" .` → analyze referenced defined template; union used keys.
- tpl:
  - If argument is literal or from default `values.yaml` and parses as template → analyze it recursively.
  - If argument can be overridden by user → mark as dynamic; warn (error in `--strict`).
- Root `.Values` usage:
  - If templates reference bare `.Values` (or `$ .Values`) treat usage as dynamic. Suppress “defined-but-not-used” claims unless `--strict` and dynamic accesses are fully resolved.

## Variable & Scope Handling
- Maintain a stack of scopes. Push on entering `If/With/Range` bodies and `define` bodies.
- Support common assignment forms:
  - `:=` in pipelines (declarations) and `=` (reassignments).
- Shadowing: inner scope bindings override outer ones.
- Data-flow is intra-template and intra-call; no cross-render execution.

## Schema Integration
- Load `values.schema.json` if present.
- Allow pluggable schema providers (e.g., generated schemas) when a chart does not ship one.
- Provide helpers to answer:
  - Is a path defined? What type (object/array/string/etc.)?
  - For objects: is `additionalProperties` false (closed object)? Enumerated `properties`?
  - For strings: enum values?
- Use schema to:
  - Resolve LiteralSet for enum-driven key selections.
  - Decide if a map’s entries are finite (closed) to mark child keys used under `range`.
  - Validate referenced-but-undefined keys.

## Handling Common Patterns
- Direct field chains: fully supported.
- index/dig with literal keys: fully supported.
- index/dig with variables:
  - If variable resolves to LiteralSet: supported.
  - Else: dynamic (warn; fail in `--strict`).
- range over lists/maps: mark container as used; optionally expand to children under strict+closed schema.
- default/required/ternaries: treat as reads of the non-default operand.
- include/template: follow call graph (with recursion guard).
- tpl: literal/default-only; override-origin tpl is dynamic.
 - Parent/child heuristics: consider a parent key effectively used if any child key is used; avoid flagging container-only keys as unused when children exist.

## Strict Mode
- `--strict` enforces:
  - All index/dig/get key operands must resolve to finite LiteralSet via literals or schema (enums/closed objects).
  - `tpl` on non-literal or override-origin sources is an error.
  - Open maps without `additionalProperties: false` cannot be claimed fully used via `range`.
- Useful for CI scenarios seeking complete static guarantees.

## Subcharts & NOTES.txt
- Subcharts: optionally traverse `charts/**/templates` and `charts/**/values.yaml` and merge schema where relevant. Flag cross-chart overrides separately.
- NOTES.txt: configurable inclusion. Default: excluded from usage calculation (optionally include via flag).

## Output & Reporting
- Categories:
  - Unused: defined in values/schema but never referenced (consider container vs leaf reporting separately).
  - Undefined: referenced by templates but not defined by values/schema.
  - Dynamic: sites where access could not be resolved statically.
- Formats: text (human), JSON (CI/machine).
- Exit codes: 0 success; non-zero based on thresholds/flags (e.g., any undefined, any unused in strict, etc.).
- Source locations: include filename:line spans from `parse.Node` positions for all findings.
- Ignore pragmas: support inline comments to suppress specific diagnostics (e.g., `# helmcheck: ignore-unused=foo.bar`).
- Explain mode: optional verbose output to show why a key is considered used (origin node/evaluator).

## CLI Flags (initial)
- `--values <path>`: path to values.yaml (default: chart root).
- `--schema <path>`: path to values.schema.json (optional).
- `--include-notes`: include `NOTES.txt` in analysis.
- `--strict`: enforce full static resolvability rules.
- `--format json|text`.
- `--fail-on unused|undefined|dynamic` (repeatable or comma-separated).

## Performance Considerations
- Parse each template file once; cache define blocks by name.
- Reuse schema lookups via memoized path resolution.
- Bounded recursion for template/include to avoid cycles.
- Analysis is linear in AST size; expect fast runs for typical charts.
 - Avoid pre-scan/quick modes; rely on deterministic AST traversal with caching for performance.

## Limitations & Future Work
- Advanced data-flow (merging branches, multi-step variable propagation) beyond simple aliasing.
- Hybrid static + instrumentation: optional render with synthetic values (from schema) to log dynamic accesses and union with static results.
- Editor integrations (LSP diagnostics) and GitHub Action packaging.
- Performance tuning for very large charts/many subcharts.

---

This document encodes the initial architecture and a pragmatic, incremental plan to quickly reach a useful state while leaving room for stricter static guarantees under `--strict`.

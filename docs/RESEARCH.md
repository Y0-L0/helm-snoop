# Survey: Existing Tools and Approaches

This document catalogs existing projects and discussions related to detecting unused/undefined Helm values and drift between values.yaml / schema and templates. Each entry notes how it differs from this project’s planned static analysis.

## Tools and Plugins

- helm-unused-values (Go, AST parse)
  - Parses templates with text/template/parse and collects direct `.Values.*` field accesses at top level. No recursive traversal into `if/with/range`, no include/define graph, no functions (index/dig), no variable aliasing, no schema integration.
  - Differs: `helm-snoop` walks the full AST (all branches), recognizes function forms, tracks variables/scopes, integrates schema (enums/closed objects), and supports strict mode.

- helm-dirty-values (Python, Helm plugin)
  - Flattens values files and regex-scans templates for `Values.<key>` references; groups unused keys by source file. No AST, no dynamic function handling, no schema awareness.
  - Differs: We avoid regex-based detection; use AST + schema to resolve dynamic access with stricter guarantees.

- helm-values-check (Go CLI)
  - Loads charts via Helm loader; detects used keys with regex for `.Values.foo.bar` and `index .Values "a" "b"`. Reports undefined and unused. No schema, no variables/scopes, no AST traversal, limited dynamic patterns.
  - Differs: We implement AST traversal, variable alias/scoping, index/dig/default/hasKey support, schema-aware resolution, and strict mode.

- helm-values-manager (Helm plugin)
  - Compares downstream values to upstream defaults; flags unsupported/redundant values and optimizes values files. Does not analyze template usage.
  - Differs: We analyze templates ↔ values/schema usage, not only values ↔ values comparisons.

- helm-highlight-unused-values (VS Code extension)
  - Highlights unused values in editors by scanning templates/values; implementation likely regex/file-scan oriented for UX.
  - Differs: We target CI/static analysis with AST + schema + strict guarantees instead of editor-only hints.

## Language Servers and Parsers

- helm-ls (language server)
  - Converts template files to YAML and delegates to yaml-language-server; generates JSON schemas for values files for completion. Not a `.Values` usage analysis.
  - Differs: We directly parse Go template AST to extract `.Values` usage and validate against values/schema.

- tree-sitter-go-template (grammar)
  - Tree-sitter grammar for Go templates, useful for editor tooling and basic parsing.
  - Differs: We rely on Go’s own template AST and add semantic interpretation (functions, scopes, schema), beyond syntax highlighting.

## Schema Tooling and Chart QA

- JSON schema generators (e.g., captainroy-hy/helm-schema-generator, SocialGouv/helm-schema, cozystack/cozyvalues-gen)
  - Generate schemas for values from defaults or CUE/processing pipelines; help enforce input constraints.
  - Differs: We consume schemas to resolve dynamic keys (enums/closed objects) and enforce `--strict` completeness; generation itself is out of scope.

- chart-testing / chart-verifier
  - General lint/testing/certification tools; do not detect unused values or template/values drift.
  - Differs: `helm-snoop` fills this specific gap.

## Community Discussions

- Helm Issue #6422 — “Warn about unused values in values file”
  - Proposal: warn/fail when values in values.yaml are not consumed by templates; strict mode suggested. Discussion points to using `values.schema.json` with `"additionalProperties": false` to reject unexpected keys in overrides. Acknowledges that many charts lack schemas; auto-generation is suggested but imperfect. Not implemented in Helm core.
  - Alignment: Our `--strict` mode and schema-aware checks provide the strict-consume behavior externally, plus undefined detection and dynamic access classification.

- Reddit: “Identify unused Helm values”
  - Context: User deploys with ArgoCD; kubeconform validates rendered manifests but Helm silently ignores unknown/unused values. Seeks a CI step to detect unused values.
  - Suggestions from comments:
    - Use values.schema.json validation with JSON Schema and set `additionalProperties: false` to reject unknown inputs; good for charts that ship schemas.
    - For third‑party charts without schemas, options include forking to add schemas or generating schemas (see arthurkoziel.com guide) as a stopgap.
    - Use Helm flow‑control functions (`default`, `fail`, `required`) to harden templates against missing inputs, though this does not detect unused keys.
    - Helm issue #6422 referenced; feature not implemented in core Helm.
  - Alignment: `helm-snoop` complements schema validation by inspecting template usage directly; works even when charts lack schemas, and can enforce strict completeness when schemas are present.

## Summary: How This Project Differs

- AST-first and comprehensive: Full Go template AST traversal (all branches), include/template call graph, functions (index/dig/get/default/hasKey), and scoped variables.
- Schema-aware: Uses enums and closed objects to resolve dynamic keys; warns on unresolved, enforces finite resolvability in `--strict`.
- tpl guardrails: Analyze literal/default-sourced `tpl`; flag override-origin dynamic templates, especially in `--strict`.
- CI-focused outputs: Unused vs undefined vs dynamic, with text/JSON reports and configurable failure conditions.

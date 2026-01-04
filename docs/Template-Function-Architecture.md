# Template Function Evaluation Architecture

## Overview

The analyzer evaluates Helm templates to discover `.Values` paths without executing them. Functions recursively evaluate arguments, tracking conservative path unions and optional structure for precise resolution.

## Core Types

**evalResult** - Return value:
```go
type evalResult struct {
    args  []string              // Literal strings (for keys, folding)
    paths []*path.Path          // .Values paths (conservative union)
    dict  map[string]*path.Path // Optional structure (nil except for dict)
}
```

**evalCtx** - Context providing `Eval(node)`, `Emit(path)`, and `WithPrefix(prefix)` for scoping

**Call** - Unevaluated function call with name and arguments

## Evaluation Model: Eval-and-Emit

Functions receive unevaluated arguments and call `ctx.Eval()` recursively. Behavior differs by runtime return type:

### Flavor 1: Complex Value Producers (dict/map/list)
Return paths for composition; **do NOT emit**. Allows further composition like `index (default .Values.a .Values.b) "key"`.

Examples: `index`, `get`, `default`, `ternary`, `coalesce`, `dict`, `merge`, `list`, `pick`, `omit`, `concat`

```go
func indexFn(ctx *evalCtx, call Call) evalResult {
    baseResult := ctx.Eval(call.Args[0])
    // ... append keys to paths ...
    return evalResult{paths: modifiedPaths}  // Return, don't emit
}
```

### Flavor 2: Simple Value Producers (string/int/bool)
Emit all argument paths immediately. Terminal operations that prevent further composition.

Examples: `quote`, `upper`, `toYaml`, `toJson`, `indent`, `nindent`, `printf`, `eq`, `ne`

```go
func quoteFn(ctx *evalCtx, call Call) evalResult {
    result := ctx.Eval(call.Args[0])
    ctx.Emit(result.paths...)  // Emit immediately
    return evalResult{args: result.args}
}
```

### Flavor 3: Control Flow (with/range)
Set paths as context prefixes; **do NOT emit**. Paths are only tracked when accessed via relative fields (`.field`) in the body.

Examples: `with`, `range`

```go
func evalWithNode(node *parse.WithNode) evalResult {
    result := e.Eval(node.Pipe)
    // Use paths as prefixes, don't emit
    restore := e.WithPrefixes(result.paths)
    e.Eval(node.List)
    restore()
    return evalResult{}
}
```

When functions like `concat` or `default` are used with control flow, they return multiple paths that all become prefixes:
```yaml
{{ range concat .Values.a .Values.b }}{{ .field }}{{ end }}
# Tracks: .Values.a.*.field and .Values.b.*.field
```

### Critical Rule: Never Combine Patterns

**Never** both emit and return paths in the same function. This would cause:
- Duplicate path tracking
- False positives in "Defined-not-used" detection
- Incorrect analysis results

**Bad Example:**
```go
func badFn(ctx *evalCtx, call Call) evalResult {
    result := ctx.Eval(call.Args[0])
    ctx.Emit(result.paths...)              // ❌ Emits
    return evalResult{paths: result.paths} // ❌ Also returns - DUPLICATE!
}
```

The correct approach: choose ONE pattern based on how the function is used:
- If used standalone (e.g., `{{ default .Values.a .Values.b }}`), ActionNode emits the returned paths
- If used with control flow (e.g., `{{ with default ... }}`), control flow uses returned paths as prefixes

## Structure Tracking

`dict` returns both conservative union and structure map for precise resolution:

```go
// {{ dict "labels" .Values.labels "context" . }}
return evalResult{
    paths: [.Values.labels],           // Conservative union
    dict:  {"labels": .Values.labels}, // Precise mapping
}
```

Enables `include` to resolve `{{ .labels.app }}` to `.Values.labels.app` instead of generic `.labels`.

Functions using structure: `index`, `get`, `include` check `dict` field for precise resolution, fall back to conservative appending if nil.

## Context Scoping

**with/range** - Prefix for relative field access:
```yaml
{{ with .Values.config }}{{ .timeout }}{{ end }}  # → .Values.config.timeout
```

**include arguments** - Control template context:
- `$` clears prefix (root)
- `.` keeps current prefix
- `.Values.foo` sets prefix to `foo`

**Built-in objects** - `.Chart`, `.Release`, `.Files`, `.Capabilities`, `.Template` are never tracked or prefixed.

## Deduplication

Parser preserves duplicates in `[]*path.Path`. Deduplication happens in output formatters, not during traversal.

## Alternatives Considered

**Subject + Usage bag** - Rejected. Emit-on-sight is simpler and aligns better with prefix stacks.

**Centralized emission** - Rejected. Harder to manage prefixes and capture predicate usages.

**Occurrence tracking** - Deferred. Formatters can correlate paths with source locations later.

See `pkg/parser/tmplFuncs.go` for implementations grouped by flavor.

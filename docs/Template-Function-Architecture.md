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

Examples: `index`, `get`, `default`, `ternary`, `coalesce`, `dict`, `merge`, `list`, `pick`, `omit`

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

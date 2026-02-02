
# Data Structures

## Path Identity
- Internal: represent paths as typed segments `Path = []Segment` where each `Segment` is either a key (string) or an index (int). This keeps intent unambiguous and supports later extensions.
- External (for map/set keys and output): render to JSON Pointer (RFC 6901) when reducing to sets/reports.
- Escaping (RFC 6901): in each token, replace `~` with `~0` and `/` with `~1`.

## Defined vs Used
- DefinedSet: map keyed by JSON Pointer (later: add a lightweight tree with `NodeKind` = Object | Array | Scalar for container context).
- Used: collect per-site `Usage` records (typed path segments, file/pos, kind, confidence), then reduce to a `map[jsonPointer][]Usage` for reporting and diffs.

## Type Mismatch Detection
- Compare typed Used paths against the Defined tree’s `NodeKind` per segment:
  - Defined container is Array, Used segment is key → TypeMismatch(ArrayExpected).
  - Defined container is Object, Used segment is index → TypeMismatch(ObjectExpected).
- Emit mismatches separately; emit JSON Pointer only for definite, resolvable paths.

## Why JSON Pointer
- Canonical, deterministic keys for definite paths (ideal for map/set comparisons).
- No wildcards; for dynamic/unknown cases, collapse to the nearest container pointer and mark as dynamic. Extensions can come later if needed.

## Examples
1. Defined: {"a":{"b":[{"c":1}]}}
   - Use `a.b[1].c` → OK → `/a/b/1/c`.
   - Use `a.b.1.c` → TypeMismatch(ObjectKeyUsedOnArray) at `/a/b`.
2. Defined: {"a":{"b":{"1":{"c":1}}}}
   - Use `a.b.1.c` → OK → `/a/b/1/c`.
   - Use `a.b[1].c` → TypeMismatch(ArrayIndexUsedOnObject) at `/a/b`.

## Suggested Minimal Types
1. Kind
   1. type kind byte // keyKind, indexKind
   2. Future: wildcardKeyKind, wildcardIndexKind, enumKind
2. Path (two slices)
   1. type Path struct { tokens []string; kinds []kind } // same length
   2. Future: enums [][]string (only for Enum segments)
3. IDs
   1. func (p Path) ID() string // RFC 6901 over definite segments
   2. Future: ExtID() including wildcards/enums (escape ~,/,* as ~0,~1,~2)
4. Builders
   1. func (p *Path) Key(s string) *Path
   2. func (p *Path) Idx(s string) *Path
   3. Future: WildcardKey(), WildcardIndex(), Enum(vals []string)
5. Match
   1. func Match(a, b Path) bool // now: Key==Key and Index==Index only
   2. Future: WildcardKey/Index match-any; Enum↔Key membership; Enum↔Enum intersect
6. Usage
   1. type Usage struct { Path Path; FileID int; Offset int; Kind UsageKind; Confidence Conf; Notes []string }
7. DefinedMeta
   1. type DefinedMeta struct { Kind NodeKind }
8. Aggregations
   1. Use `ID()` for map keys; future: carry dynamic metadata for wildcards/enums

## Map vs List of Structs
1. Extraction produces a list of Usage structs (detailed, per-site).
2. Reporting reduces to a map keyed by canonical path (dedup + collapse rules).
3. This separates concerns: extraction = detailed, reduction = summarized.

## Path Comparison Flavors

The analyzer uses three distinct comparison operations:

### 1. Strict (`Compare`)
Exact equality for sorting. Compares tokens AND kinds position by position.
- Used in: `sort.Sort()`, duplicate removal
- Example: `/foo` (keyKind) ≠ `/foo` (indexKind)

### 2. Subsumption (`subsumes`)
One-way implication for removing redundant paths.
- `/foo/*` subsumes `/foo` (wildcard implies base exists)
- `/foo/*` does NOT subsume `/foo/bar` (different constraints)
- Used in: `SortDedup()`

### 3. Loose (`EqualLoose`)
Bidirectional matching for joining definitions with usages.
- anyKind matches keyKind/indexKind
- Terminal wildcards match descendants: `/foo/*` matches `/foo/bar`
- Used in: `MergeJoinLoose()`

**Key distinction**: Subsumption vs Loose
- `/foo/*` and `/foo/bar` match loosely (for joining)
- But are NOT redundant (both needed in "undefined" reports)
  - `/foo/*` = need ANY children
  - `/foo/bar` = need THIS SPECIFIC field

| Operation | Function | Symmetric | Purpose |
|-----------|----------|-----------|---------|
| Strict | `Compare()` | Yes | Sorting |
| Subsumption | `subsumes()` | No | Deduplication |
| Loose | `EqualLoose()` | Yes | Joining defs↔usages |

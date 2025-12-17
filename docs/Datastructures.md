
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
1. SegKind
   1. type SegKind int // Key, Index, Wildcard, EnumSet (start with Key, Index)
2. Segment
   1. type Segment struct { Kind SegKind; S string; I int; Set []string }
3. Path
   1. type Path []Segment
4. Canonical ID
   1. func CanonicalID(Path) string // JSON Pointer for definite; extend later
5. Usage
   1. type Usage struct { Path Path; FileID int; Offset int; Kind UsageKind; Confidence Conf; Notes []string }
6. DefinedMeta
   1. type DefinedMeta struct { Kind NodeKind }
7. Aggregations
   1. `map[string][]Usage`; `map[string]DefinedMeta`

## Map vs List of Structs
1. Extraction produces a list of Usage structs (detailed, per-site).
2. Reporting reduces to a map keyed by canonical path (dedup + collapse rules).
3. This separates concerns: extraction = detailed, reduction = summarized.

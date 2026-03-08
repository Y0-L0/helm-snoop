package tplparser

import (
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

type evalResult struct {
	// args contains literal string values extracted from the node.
	// Used for literal folding and as keys for path synthesis (index, get).
	args []string

	// paths contains the union of all .Values paths discovered during evaluation.
	paths []*vpath.Path

	// dict provides structure tracking for dict literals.
	// Maps literal keys to nested evalResults, supporting arbitrary nesting.
	// Only populated by the dict function; nil otherwise.
	// Functions like index, get, and include can use this for precise resolution.
	dict map[string]evalResult

	// dictLits tracks literal values in dict (not paths).
	dictLits map[string]string
}

// hasDict reports whether this result carries dict structure.
func (r evalResult) hasDict() bool {
	return r.dict != nil || r.dictLits != nil
}

// resolveFields walks through dict structure to resolve a field chain.
// If the first field matches a dict key, resolution continues recursively.
// Otherwise, remaining fields are appended to all paths.
func (r evalResult) resolveFields(fields []string) evalResult {
	if len(fields) == 0 {
		return r
	}

	if r.dict != nil {
		if inner, ok := r.dict[fields[0]]; ok {
			return inner.resolveFields(fields[1:])
		}
	}

	// Strip .Values when base path is root — the root context ($) already
	// represents the chart root, so "Values" is not a real path segment.
	if len(r.paths) > 0 && r.paths[0].ID() == "." && fields[0] == "Values" {
		fields = fields[1:]
		if len(fields) == 0 {
			return evalResult{}
		}
	}

	// Leaf: append remaining fields to paths.
	var paths []*vpath.Path
	for _, p := range r.paths {
		resolved := *p
		for _, f := range fields {
			resolved = resolved.WithKey(f)
		}
		paths = append(paths, &resolved)
	}
	return evalResult{paths: paths}
}

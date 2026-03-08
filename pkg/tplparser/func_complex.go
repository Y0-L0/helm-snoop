package tplparser

import (
	"log/slog"

	"github.com/y0-l0/helm-snoop/internal/assert"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

// ==============================================================================
// FLAVOR 1: COMPLEX VALUE PRODUCERS
// ==============================================================================

// dictFn builds both conservative union AND structure tracking.
func dictFn(ctx *evalCtx, call Call) evalResult {
	dict := make(map[string]evalResult)
	dictLits := make(map[string]string)
	var allPaths []*vpath.Path

	for i := 0; i+1 < len(call.Args); i += 2 {
		keyResult := ctx.Eval(call.Args[i])
		valResult := ctx.Eval(call.Args[i+1])

		allPaths = append(allPaths, valResult.paths...)

		if len(keyResult.args) != 1 {
			continue
		}
		key := keyResult.args[0]

		if len(valResult.paths) > 0 || valResult.dict != nil {
			dict[key] = valResult
		}

		if len(valResult.args) > 0 {
			dictLits[key] = valResult.args[0]
		}
	}

	return evalResult{
		paths:    allPaths,
		dict:     orNil(dict),
		dictLits: orNil(dictLits),
	}
}

// indexFn builds child paths by appending keys.
func indexFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) < 2 {
		slog.Warn("index: need at least 2 args", "count", len(call.Args))
		assert.Must("index: need at least 2 args")
		return evalResult{}
	}

	baseResult := ctx.Eval(call.Args[0])

	// Check for dict structure
	if baseResult.dict != nil {
		// Resolve through dict for each key arg, chaining results.
		result := baseResult
		for _, arg := range call.Args[1:] {
			keyResult := ctx.Eval(arg)
			if len(keyResult.args) == 1 {
				result = result.resolveFields(keyResult.args)
			}
		}
		return result
	}

	// Fall back to conservative: append keys to all paths
	var keys []string
	for _, arg := range call.Args[1:] {
		keyResult := ctx.Eval(arg)
		keys = append(keys, keyResult.args...)
	}

	var modifiedPaths []*vpath.Path
	for _, base := range baseResult.paths {
		p := *base
		for _, key := range keys {
			p = p.WithAny(key)
		}
		modifiedPaths = append(modifiedPaths, &p)
	}

	return evalResult{paths: modifiedPaths}
}

// getFn is like index but with exactly 2 args.
func getFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) != 2 {
		slog.Warn("get: need exactly 2 args", "count", len(call.Args))
		assert.Must("get: need exactly 2 args")
		return evalResult{}
	}

	baseResult := ctx.Eval(call.Args[0])
	keyResult := ctx.Eval(call.Args[1])

	// Check for dict structure
	if baseResult.dict != nil && len(keyResult.args) == 1 {
		return baseResult.resolveFields(keyResult.args)
	}

	// Fall back to conservative
	if len(keyResult.args) != 1 {
		slog.Warn("get: key must be exactly one literal", "count", len(keyResult.args))
		assert.Must("get: key must be exactly one literal")
		return evalResult{}
	}

	var modifiedPaths []*vpath.Path
	for _, base := range baseResult.paths {
		p := *base
		p = p.WithAny(keyResult.args[0])
		modifiedPaths = append(modifiedPaths, &p)
	}

	return evalResult{paths: modifiedPaths}
}

// mergeFn merges dict structures from all arguments (first key wins).
func mergeFn(ctx *evalCtx, call Call) evalResult {
	return mergeArgs(ctx, call, false)
}

// mergeOverwriteFn merges dict structures (last key wins).
func mergeOverwriteFn(ctx *evalCtx, call Call) evalResult {
	return mergeArgs(ctx, call, true)
}

func mergeArgs(ctx *evalCtx, call Call, overwrite bool) evalResult {
	results := make([]evalResult, len(call.Args))
	var allPaths []*vpath.Path
	hasDict := false

	for i, arg := range call.Args {
		results[i] = ctx.Eval(arg)
		allPaths = append(allPaths, results[i].paths...)
		hasDict = hasDict || results[i].hasDict()
	}

	if !hasDict {
		return evalResult{paths: allPaths}
	}

	merged := make(map[string]evalResult)
	mergedLits := make(map[string]string)
	for _, r := range results {
		mergeDictInto(merged, r.dict, overwrite)
		mergeLitsInto(mergedLits, r.dictLits, overwrite)
	}

	return evalResult{
		paths:    allPaths,
		dict:     orNil(merged),
		dictLits: orNil(mergedLits),
	}
}

func mergeDictInto(dst map[string]evalResult, src map[string]evalResult, overwrite bool) {
	for k, v := range src {
		if _, exists := dst[k]; !exists || overwrite {
			dst[k] = v
		}
	}
}

func mergeLitsInto(dst map[string]string, src map[string]string, overwrite bool) {
	for k, v := range src {
		if _, exists := dst[k]; !exists || overwrite {
			dst[k] = v
		}
	}
}

func orNil[M ~map[K]V, K comparable, V any](m M) M {
	if len(m) == 0 {
		return nil
	}
	return m
}

// defaultFn unions all argument paths and returns them.
func defaultFn(ctx *evalCtx, call Call) evalResult {
	var allPaths vpath.Paths
	for _, arg := range call.Args {
		result := ctx.Eval(arg)
		allPaths = append(allPaths, result.paths...)
	}

	return evalResult{paths: allPaths}
}

package parser

import (
	"log/slog"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// ==============================================================================
// FLAVOR 1: COMPLEX VALUE PRODUCERS
// ==============================================================================

// dictFn builds both conservative union AND structure tracking
func dictFn(ctx *evalCtx, call Call) evalResult {
	dict := make(map[string]*path.Path)
	dictLits := make(map[string]string)
	var allPaths []*path.Path

	for i := 0; i+1 < len(call.Args); i += 2 {
		keyResult := ctx.Eval(call.Args[i])
		valResult := ctx.Eval(call.Args[i+1])

		allPaths = append(allPaths, valResult.paths...)

		if len(keyResult.args) != 1 {
			continue
		}
		key := keyResult.args[0]

		if len(valResult.paths) > 0 {
			dict[key] = valResult.paths[0]
		}

		if len(valResult.args) > 0 {
			dictLits[key] = valResult.args[0]
		}
	}

	return evalResult{
		paths:    allPaths,
		dict:     dict,
		dictLits: dictLits,
	}
}

// indexFn builds child paths by appending keys
func indexFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) < 2 {
		slog.Warn("index: need at least 2 args", "count", len(call.Args))
		Must("index: need at least 2 args")
		return evalResult{}
	}

	baseResult := ctx.Eval(call.Args[0])

	// Check for dict structure
	if baseResult.dict != nil {
		var resolvedPaths []*path.Path

		for _, arg := range call.Args[1:] {
			keyResult := ctx.Eval(arg)
			for _, key := range keyResult.args {
				if p, ok := baseResult.dict[key]; ok {
					resolvedPaths = append(resolvedPaths, p)
				}
			}
		}

		return evalResult{paths: resolvedPaths}
	}

	// Fall back to conservative: append keys to all paths
	var keys []string
	for _, arg := range call.Args[1:] {
		keyResult := ctx.Eval(arg)
		keys = append(keys, keyResult.args...)
	}

	var modifiedPaths []*path.Path
	for _, base := range baseResult.paths {
		p := *base
		for _, key := range keys {
			p = p.WithAny(key)
		}
		modifiedPaths = append(modifiedPaths, &p)
	}

	return evalResult{paths: modifiedPaths}
}

// getFn is like index but with exactly 2 args
func getFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) != 2 {
		slog.Warn("get: need exactly 2 args", "count", len(call.Args))
		Must("get: need exactly 2 args")
		return evalResult{}
	}

	baseResult := ctx.Eval(call.Args[0])
	keyResult := ctx.Eval(call.Args[1])

	// Check for dict structure
	if baseResult.dict != nil && len(keyResult.args) == 1 {
		if p, ok := baseResult.dict[keyResult.args[0]]; ok {
			return evalResult{paths: []*path.Path{p}}
		}
		return evalResult{}
	}

	// Fall back to conservative
	if len(keyResult.args) != 1 {
		slog.Warn("get: key must be exactly one literal", "count", len(keyResult.args))
		Must("get: key must be exactly one literal")
		return evalResult{}
	}

	var modifiedPaths []*path.Path
	for _, base := range baseResult.paths {
		p := *base
		p = p.WithAny(keyResult.args[0])
		modifiedPaths = append(modifiedPaths, &p)
	}

	return evalResult{paths: modifiedPaths}
}

// defaultFn unions all argument paths and returns them.
func defaultFn(ctx *evalCtx, call Call) evalResult {
	var allPaths path.Paths
	for _, arg := range call.Args {
		result := ctx.Eval(arg)
		allPaths = append(allPaths, result.paths...)
	}

	return evalResult{paths: allPaths}
}

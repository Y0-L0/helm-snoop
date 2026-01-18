package parser

import (
	"log/slog"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

func noopStrings(_ *evalCtx, _ Call) evalResult {
	return evalResult{}
}

// ==============================================================================
// FLAVOR 2: SIMPLE VALUE PRODUCERS (emit paths immediately)
// ==============================================================================

// quoteFn emits arg paths and returns strings if available
func quoteFn(ctx *evalCtx, call Call) evalResult {
	var allStrings []string

	for _, arg := range call.Args {
		result := ctx.Eval(arg)

		// Emit paths immediately
		ctx.Emit(call.Node.Position(), result.paths...)

		allStrings = append(allStrings, result.args...)
	}

	return evalResult{args: allStrings}
}

// unaryPassThroughFn emits the arg path if present, returns literal strings
func unaryPassThroughFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) == 0 {
		return evalResult{}
	}

	result := ctx.Eval(call.Args[0])

	ctx.Emit(call.Node.Position(), result.paths...)

	return evalResult{args: result.args}
}

// unarySerializeFn emits the arg path with a terminal wildcard (for toYaml, toJson, etc.)
func unarySerializeFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) == 0 {
		return evalResult{}
	}

	result := ctx.Eval(call.Args[0])

	for _, p := range result.paths {
		wildcardPath := p.WithWildcard()
		ctx.Emit(call.Node.Position(), &wildcardPath)
	}

	return evalResult{args: result.args}
}

// emitArgsNoResultFn evaluates all args and emits any .Values paths found
func emitArgsNoResultFn(ctx *evalCtx, call Call) evalResult {
	for _, arg := range call.Args {
		result := ctx.Eval(arg)
		ctx.Emit(call.Node.Position(), result.paths...)
	}

	return evalResult{}
}

// omitPickFn handles omit/pick functions - evaluates all args but returns only the first arg's paths
// This allows the result to be piped to other functions like toYaml
func omitPickFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) == 0 {
		return evalResult{}
	}

	result := ctx.Eval(call.Args[0])

	for i := 1; i < len(call.Args); i++ {
		_ = ctx.Eval(call.Args[i])
	}

	return evalResult{paths: result.paths, args: result.args}
}

// concatFn evaluates all args and returns all paths (for concat which merges lists)
func concatFn(ctx *evalCtx, call Call) evalResult {
	var allPaths path.Paths

	for _, arg := range call.Args {
		result := ctx.Eval(arg)
		allPaths = append(allPaths, result.paths...)
	}

	return evalResult{paths: allPaths}
}

// binaryEvalFn evaluates the first 2 args and emits any .Values paths found
// Used for comparison, string manipulation, and type checking functions
func binaryEvalFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) != 2 {
		slog.Warn("binary function requires exactly 2 arguments", "count", len(call.Args))
		Must("binary function requires exactly 2 arguments")
	}

	if len(call.Args) < 2 {
		return evalResult{}
	}

	// Only evaluate the first 2 arguments
	for _, arg := range call.Args[:2] {
		result := ctx.Eval(arg)
		ctx.Emit(call.Node.Position(), result.paths...)
	}

	return evalResult{}
}

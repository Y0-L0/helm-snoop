package parser

import "log/slog"

// ==============================================================================
// FLAVOR 2: SIMPLE VALUE PRODUCERS (emit paths immediately)
// ==============================================================================

// quoteFn emits arg paths and returns strings if available
func quoteFn(ctx *evalCtx, call Call) evalResult {
	var allStrings []string

	for _, arg := range call.Args {
		result := ctx.Eval(arg)

		// Emit paths immediately
		ctx.Emit(result.paths...)

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

	// Emit paths
	ctx.Emit(result.paths...)

	return evalResult{args: result.args}
}

// unarySerializeFn emits the arg path with a terminal wildcard (for toYaml, toJson, etc.)
// These functions serialize the entire subtree, so we add /* to indicate all descendants are accessed
func unarySerializeFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) == 0 {
		return evalResult{}
	}

	result := ctx.Eval(call.Args[0])

	// Add terminal wildcard to each path
	for _, p := range result.paths {
		wildcardPath := p.WithWildcard()
		ctx.Emit(&wildcardPath)
	}

	return evalResult{args: result.args}
}

// emitArgsNoResultFn evaluates all args and emits any .Values paths found
func emitArgsNoResultFn(ctx *evalCtx, call Call) evalResult {
	for _, arg := range call.Args {
		result := ctx.Eval(arg)
		ctx.Emit(result.paths...)
	}

	return evalResult{}
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
		ctx.Emit(result.paths...)
	}

	return evalResult{}
}

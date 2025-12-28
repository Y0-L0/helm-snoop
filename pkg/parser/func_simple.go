package parser

// ==============================================================================
// FLAVOR 2: SIMPLE VALUE PRODUCERS (emit paths immediately)
// ==============================================================================

// quoteFn emits arg paths and returns strings if available
func quoteFn(ctx *evalCtx, call Call) evalResult {
	var allStrings []string

	for _, arg := range call.Args {
		result := ctx.Eval(arg)

		// Emit paths immediately
		for _, p := range result.paths {
			ctx.Emit(p)
		}

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
	for _, p := range result.paths {
		ctx.Emit(p)
	}

	return evalResult{args: result.args}
}

// emitArgsNoResultFn evaluates all args and emits any .Values paths found
func emitArgsNoResultFn(ctx *evalCtx, call Call) evalResult {
	for _, arg := range call.Args {
		result := ctx.Eval(arg)
		for _, p := range result.paths {
			ctx.Emit(p)
		}
	}

	return evalResult{}
}

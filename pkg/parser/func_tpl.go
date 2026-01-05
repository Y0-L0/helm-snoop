package parser

import (
	"log/slog"
)

// ==============================================================================
// NOT IMPLEMENTED / PLACEHOLDERS
// ==============================================================================

func tplFn(ctx *evalCtx, call Call) evalResult {
	slog.Warn("tpl partially implemented (ignoring context argument)")
	Must("tpl partially implemented (ignoring context argument)")

	if len(call.Args) < 1 {
		return evalResult{}
	}

	result := ctx.Eval(call.Args[0])

	return result
}

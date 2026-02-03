package parser

import (
	"log/slog"

	"github.com/y0-l0/helm-snoop/internal/assert"
)

// ==============================================================================
// NOT IMPLEMENTED / PLACEHOLDERS
// ==============================================================================

func tplFn(ctx *evalCtx, call Call) evalResult {
	slog.Info("tpl partially implemented (ignoring context argument)")
	assert.Must("tpl partially implemented (ignoring context argument)")

	if len(call.Args) < 1 {
		return evalResult{}
	}

	result := ctx.Eval(call.Args[0])

	return result
}

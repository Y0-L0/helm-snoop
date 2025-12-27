package parser

import (
	"log/slog"
)

// include/tpl remain out of scope; keep them as strict-not-implemented.
func includeFn(ctx *FnCtx, call Call) []string {
	// Expect template name as first arg (literal string)
	if len(call.Args) == 0 {
		slog.Warn("include: missing template name")
		Must("include: missing template name")
		return nil
	}
	lits := ctx.EvalNode(call.Args[0]).Strings
	if len(lits) != 1 {
		slog.Warn("include: invalid name literal", "len", len(lits))
		Must("include: invalid name literal")
		return nil
	}
	name := lits[0]
	if ctx.a == nil || ctx.a.idx == nil {
		slog.Warn("include: no template index")
		Must("include: no template index")
		return nil
	}
	def, ok := ctx.a.idx.get(name)
	if !ok {
		slog.Warn("include: unknown template name", "name", name)
		Must("include: unknown template name")
		return nil
	}
	defer ctx.a.withIncludeScope(name)()
	oldTree := ctx.a.tree
	ctx.a.tree = def.tree
	ctx.a.collect(def.root)
	ctx.a.tree = oldTree
	return nil
}
func tplFn(_ *FnCtx, _ Call) []string {
	slog.Warn("tpl not implemented")
	Must("tpl not implemented")
	return nil
}

// makeNotImplementedFn logs and fails in strict mode.
func makeNotImplementedFn(name string) templFunc {
	return func(_ *FnCtx, _ Call) []string {
		slog.Warn("template function not implemented", "name", name)
		Must("template function not implemented: " + name)
		return nil
	}
}

func noopStrings(_ *FnCtx, _ Call) []string { return nil }

// emitArgsNoResultFn evaluates all args and emits any .Values paths found.
func emitArgsNoResultFn(ctx *FnCtx, call Call) []string {
	for _, arg := range call.Args {
		ev := ctx.EvalNode(arg)
		if ev.Path != nil {
			ctx.Emit(ev.Path)
		}
	}
	return nil
}

// unaryPassThroughFn emits the arg path if present, and returns literal string if available.
func unaryPassThroughFn(ctx *FnCtx, call Call) []string {
	if len(call.Args) == 0 && !call.Piped {
		return nil
	}
	var ev Eval
	if len(call.Args) > 0 {
		ev = ctx.EvalNode(call.Args[0])
	} else if call.Piped {
		if call.InputPath != nil {
			ctx.Emit(call.InputPath)
		}
		return call.Input
	}
	if ev.Path != nil {
		ctx.Emit(ev.Path)
	}
	return ev.Strings
}

// getFn: base must be a path; key must be a literal string; emit child path.
func getFn(ctx *FnCtx, call Call) []string {
	if len(call.Args) < 1 {
		slog.Warn("get: invalid template arg count", "args_len", len(call.Args))
		Must("get: invalid template")
		return nil
	}
	base := ctx.EvalNode(call.Args[0]).Path
	if base == nil {
		slog.Warn("get: base is not a path")
		Must("get: base is not a path")
		return nil
	}
	var keys []string
	if len(call.Args) >= 2 {
		keys = ctx.EvalNode(call.Args[1]).Strings
	} else if call.Piped && len(call.Input) > 0 {
		keys = call.Input
	}
	if len(keys) != 1 {
		slog.Warn("get: key must be exactly one literal", "len", len(keys))
		Must("get: key length != 1")
		return nil
	}
	p := *base
	p = p.WithAny(keys[0])
	ctx.Emit(&p)
	return nil
}

// indexFn: base must be a path; subsequent args must be literal strings; emit child path.
func indexFn(ctx *FnCtx, call Call) []string {
	if len(call.Args) < 1 {
		slog.Warn("index: invalid template arg count", "args_len", len(call.Args))
		Must("index: invalid template")
		return nil
	}
	// base may be provided as first arg, or piped as a previously synthesized path.
	base := ctx.EvalNode(call.Args[0]).Path
	baseIdx := 0
	if base == nil && call.Piped && call.InputPath != nil {
		base = call.InputPath
		baseIdx = -1
	}
	if base == nil {
		baseIdx = -1
		for i, n := range call.Args {
			if b := ctx.EvalNode(n).Path; b != nil {
				base = b
				baseIdx = i
				break
			}
		}
	}
	if base == nil {
		slog.Warn("index: base is not a path")
		Must("index: base is not a path")
		return nil
	}
	p := *base
	start := baseIdx + 1
	if start < 1 {
		start = 1
	}
	for i := start; i < len(call.Args); i++ {
		lit := ctx.EvalNode(call.Args[i]).Strings
		if len(lit) != 1 {
			slog.Warn("index: key must be exactly one literal")
			Must("index: key must be exactly one literal")
			return nil
		}
		p = p.WithAny(lit[0])
	}
	// keys may also come from piped input strings
	if call.Piped && len(call.Input) > 0 {
		for _, lit := range call.Input {
			p = p.WithAny(lit)
		}
	}
	ctx.Emit(&p)
	return nil
}

// defaultFn: emit any arg paths; returns no strings.
func defaultFn(ctx *FnCtx, call Call) []string {
	for _, arg := range call.Args {
		ev := ctx.EvalNode(arg)
		if ev.Path != nil {
			ctx.Emit(ev.Path)
		}
	}
	return nil
}

// no coalesceFn implementation yet (intentionally omitted)

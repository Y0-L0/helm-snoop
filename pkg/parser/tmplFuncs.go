package parser

import "github.com/y0-l0/helm-snoop/pkg/path"

func includeFn(...interface{}) interface{} { panic("not implemented") }
func tplFn(...interface{}) interface{}     { panic("not implemented") }
func noopFn(...interface{}) interface{}    { return nil }

// getFn appends a literal key (unknown kind) to an absolute .Values path.
func getFn(args ...interface{}) interface{} {
	if len(args) != 2 {
		panic("not implemented / invalid template")
	}
	base, ok := args[0].(*path.Path)
	if !ok {
		panic("not implemented")
	}
	key, ok := args[1].(LiteralSet)
	if !ok {
		panic("not implemented")
	}
	if len(key.Values) != 1 {
		panic("not implemented")
	}
	p := *base
	p = p.WithAny(key.Values[0])
	return &p
}

// indexFn appends one or more literal keys (unknown kind) to an absolute .Values path.
func indexFn(args ...interface{}) interface{} {
	if len(args) < 2 {
		panic("not implemented / invalid template")
	}
	base, ok := args[0].(*path.Path)
	if !ok {
		panic("not implemented")
	}
	p := *base
	for _, a := range args[1:] {
		lit, ok := a.(LiteralSet)
		if !ok {
			panic("not implemented")
		}
		if len(lit.Values) != 1 {
			panic("not implemented")
		}
		p = p.WithAny(lit.Values[0])
	}
	return &p
}

// Analysis-aware helpers used by our evaluator.
func defaultFn(args ...interface{}) interface{} {
	if len(args) == 0 {
		return nil
	}
	// Prefer returning a .Values path if present (order-agnostic for piping)
	for i := len(args) - 1; i >= 0; i-- {
		if ap, ok := args[i].(*path.Path); ok {
			return ap
		}
	}
	return nil
}

func passthrough1Fn(args ...interface{}) interface{} {
	if len(args) == 0 {
		return nil
	}
	return args[0]
}

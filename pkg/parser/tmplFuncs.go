package parser

import (
	"log/slog"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

func includeFn(...interface{}) interface{} {
	slog.Warn("include not implemented")
	must("include not implemented")
	return nil
}
func tplFn(...interface{}) interface{} {
	slog.Warn("tpl not implemented")
	must("tpl not implemented")
	return nil
}

// makeNotImplementedFn returns a templFunc that logs the function name
// and calls must() to fail in strict mode. Use when we want to surface
// unsupported helper usage instead of silently no-oping.
func makeNotImplementedFn(name string) templFunc {
	return func(...interface{}) interface{} {
		slog.Warn("template function not implemented", "name", name)
		must("template function not implemented: " + name)
		return nil
	}
}
func noopFn(...interface{}) interface{} { return nil }

// getFn appends a literal key (unknown kind) to a .Values Path.
func getFn(args ...interface{}) interface{} {
	if len(args) != 2 {
		slog.Warn("get: invalid template arg count", "args_len", len(args))
		must("get: invalid template")
		return nil
	}
	base, ok := args[0].(*path.Path)
	if !ok {
		slog.Warn("get: base is not a path", "base", args[0])
		must("get: base is not a path")
		return nil
	}
	key, ok := args[1].(KeySet)
	if !ok {
		slog.Warn("get: key is not a KeySet", "key", args[1])
		must("get: key is not a KeySet")
		return nil
	}
	if len(key) != 1 {
		slog.Warn("get: key length != 1", "len", len(key))
		must("get: key length != 1")
		return nil
	}
	p := *base
	p = p.WithAny(key[0])
	return &p
}

// indexFn appends one or more literal keys (unknown kind) to an absolute .Values path.
func indexFn(args ...interface{}) interface{} {
	if len(args) < 2 {
		slog.Warn("index: invalid template arg count", "args_len", len(args))
		must("index: invalid template")
		return nil
	}
	base, ok := args[0].(*path.Path)
	if !ok {
		slog.Warn("index: base is not a path", "base", args[0])
		must("index: base is not a path")
		return nil
	}
	p := *base
	for _, a := range args[1:] {
		lit, ok := a.(KeySet)
		if !ok {
			slog.Warn("index: key is not a KeySet", "arg", a)
			must("index: key is not a KeySet")
			return nil
		}
		if len(lit) != 1 {
			slog.Warn("index: key length != 1", "len", len(lit))
			must("index: key length != 1")
			return nil
		}
		p = p.WithAny(lit[0])
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

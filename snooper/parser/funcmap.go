package parser

import (
	sprig "github.com/Masterminds/sprig/v3"
)

// newFuncMap returns a tolerant function map for parsing templates.
// For out-of-scope helpers, we explicitly panic if ever invoked.
func newFuncMap() map[string]interface{} {
	funcMap := sprig.TxtFuncMap()

	// Named stubs so we can add real logic later.
	funcMap["include"] = includeFn
	funcMap["tpl"] = tplFn
	// Note: "template" and "define" are Go template actions, not functions,
	// so adding them to the FuncMap has no effect on parsing; left out on purpose.

	// Analysis-aware overrides
	funcMap["default"] = defaultFn
	funcMap["quote"] = passthrough1Fn
	funcMap["upper"] = passthrough1Fn
	funcMap["lower"] = passthrough1Fn

	funcMap["toToml"] = noopFn
	funcMap["toYaml"] = noopFn
	funcMap["fromYaml"] = noopFn
	funcMap["fromYamlArray"] = noopFn
	funcMap["toJson"] = noopFn
	funcMap["fromJson"] = noopFn
	funcMap["fromJsonArray"] = noopFn
	funcMap["required"] = noopFn
	funcMap["lookup"] = noopFn
	funcMap["and"] = noopFn
	funcMap["call"] = noopFn
	funcMap["html"] = noopFn
	funcMap["index"] = noopFn
	funcMap["slice"] = noopFn
	funcMap["js"] = noopFn
	funcMap["len"] = noopFn
	funcMap["not"] = noopFn
	funcMap["or"] = noopFn
	funcMap["print"] = noopFn
	funcMap["printf"] = noopFn
	funcMap["println"] = noopFn
	funcMap["urlquery"] = noopFn
	funcMap["eq"] = noopFn
	funcMap["ge"] = noopFn
	funcMap["gt"] = noopFn
	funcMap["le"] = noopFn
	funcMap["lt"] = noopFn
	funcMap["ne"] = noopFn

	return funcMap
}

func includeFn(...interface{}) interface{} { panic("not implemented") }
func tplFn(...interface{}) interface{}     { panic("not implemented") }
func noopFn(...interface{}) interface{}    { return nil }

// Analysis-aware helpers used by our evaluator.
func defaultFn(args ...interface{}) interface{} {
	if len(args) == 0 {
		return nil
	}
	return args[len(args)-1]
}

func passthrough1Fn(args ...interface{}) interface{} {
	if len(args) == 0 {
		return nil
	}
	return args[0]
}

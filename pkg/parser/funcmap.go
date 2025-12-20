package parser

import (
	"fmt"
	"log/slog"

	sprig "github.com/Masterminds/sprig/v3"
)

type templFunc = func(...interface{}) interface{}

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
	funcMap["get"] = getFn
	funcMap["index"] = indexFn

	funcMap["and"] = noopFn
	funcMap["call"] = noopFn
	funcMap["contains"] = noopFn
	funcMap["contains"] = noopFn
	funcMap["dict"] = noopFn
	funcMap["eq"] = noopFn
	funcMap["fail"] = noopFn
	funcMap["fromJson"] = noopFn
	funcMap["fromJsonArray"] = noopFn
	funcMap["fromToml"] = noopFn
	funcMap["fromYaml"] = noopFn
	funcMap["fromYamlArray"] = noopFn
	funcMap["ge"] = noopFn
	funcMap["getHostByName"] = noopFn
	funcMap["gt"] = noopFn
	funcMap["html"] = noopFn
	funcMap["indent"] = noopFn
	funcMap["js"] = noopFn
	funcMap["le"] = noopFn
	funcMap["len"] = noopFn
	funcMap["lookup"] = noopFn
	funcMap["lt"] = noopFn
	funcMap["merge"] = noopFn
	funcMap["mustToJson"] = noopFn
	funcMap["mustToYaml"] = noopFn
	funcMap["ne"] = noopFn
	funcMap["nindent"] = noopFn
	funcMap["nindent"] = noopFn
	funcMap["not"] = noopFn
	funcMap["omit"] = noopFn
	funcMap["or"] = noopFn
	funcMap["print"] = noopFn
	funcMap["printf"] = noopFn
	funcMap["println"] = noopFn
	funcMap["replace"] = noopFn
	funcMap["required"] = noopFn
	funcMap["semverCompare"] = noopFn
	funcMap["sha256sum"] = noopFn
	funcMap["slice"] = noopFn
	funcMap["toJson"] = noopFn
	funcMap["toToml"] = noopFn
	funcMap["toYaml"] = noopFn
	funcMap["toYamlPretty"] = noopFn
	funcMap["trimSuffix"] = noopFn
	funcMap["trunc"] = noopFn
	funcMap["urlquery"] = noopFn

	// Ensure all functions in the map conform to templFunc so our evaluator
	// can dispatch uniformly without signature panics.
	for k, v := range funcMap {
		if _, ok := v.(templFunc); !ok {
			funcMap[k] = noopFn
		}
	}
	return funcMap
}

func getTemplateFunction(name string) templFunc {
	value, ok := templFuncMap[name]
	if !ok {
		slog.Warn("unknown template function name", "name", name)
		panic(fmt.Sprintf("unknown template function name: %s", name))
	}
	function, ok := value.(templFunc)
	if !ok {
		slog.Warn("invalid template function signature", "name", name, "value", value)
		panic(fmt.Sprintf("invalid template function signature for function name: %s", name))
	}
	return function
}

// Build once and reuse to avoid recreating the map frequently.
var templFuncMap = newFuncMap()

package parser

import (
	"fmt"
	"log/slog"
)

type templFunc = func(...interface{}) interface{}

// newFuncMap returns a tolerant function map for parsing templates.
// For out-of-scope helpers, we explicitly panic if ever invoked.
var templFuncMap = map[string]interface{}{
	// Analysis-aware real handlers
	"include": includeFn,
	"tpl":     tplFn,
	"default": defaultFn,
	"get":     getFn,
	"index":   indexFn,

	// Pass-through helpers (preserve .Values usage)
	"indent":   passthrough1Fn,
	"lower":    passthrough1Fn,
	"nindent":  passthrough1Fn,
	"quote":    passthrough1Fn,
	"required": passthrough1Fn,
	"upper":    passthrough1Fn,

	// TODO: Should return wildcard key kinds in the future
	"toYaml":       passthrough1Fn,
	"mustToYaml":   passthrough1Fn,
	"toYamlPretty": passthrough1Fn,
	"toJson":       passthrough1Fn,
	"mustToJson":   passthrough1Fn,
	"toToml":       passthrough1Fn,

	// Data conversion: ignored in analysis
	"fromJson":      noopFn,
	"fromJsonArray": noopFn,
	"fromYaml":      noopFn,
	"fromYamlArray": noopFn,

	// Common collection/string helpers (ignored for now)
	"and":                    makeNotImplementedFn("and"),
	"append":                 makeNotImplementedFn("append"),
	"b64dec":                 makeNotImplementedFn("b64dec"),
	"b64enc":                 makeNotImplementedFn("b64enc"),
	"call":                   makeNotImplementedFn("call"),
	"cat":                    makeNotImplementedFn("cat"),
	"concat":                 makeNotImplementedFn("concat"),
	"contains":               makeNotImplementedFn("contains"),
	"derivePassword":         makeNotImplementedFn("derivePassword"),
	"dict":                   makeNotImplementedFn("dict"),
	"empty":                  makeNotImplementedFn("empty"),
	"eq":                     makeNotImplementedFn("eq"),
	"first":                  makeNotImplementedFn("first"),
	"ge":                     makeNotImplementedFn("ge"),
	"gt":                     makeNotImplementedFn("gt"),
	"has":                    makeNotImplementedFn("has"),
	"hasKey":                 makeNotImplementedFn("hasKey"),
	"html":                   makeNotImplementedFn("html"),
	"int":                    makeNotImplementedFn("int"),
	"join":                   makeNotImplementedFn("join"),
	"js":                     makeNotImplementedFn("js"),
	"keys":                   makeNotImplementedFn("keys"),
	"kindIs":                 makeNotImplementedFn("kindIs"),
	"le":                     makeNotImplementedFn("le"),
	"len":                    makeNotImplementedFn("len"),
	"list":                   makeNotImplementedFn("list"),
	"lt":                     makeNotImplementedFn("lt"),
	"merge":                  makeNotImplementedFn("merge"),
	"mergeOverwrite":         makeNotImplementedFn("mergeOverwrite"),
	"ne":                     makeNotImplementedFn("ne"),
	"not":                    makeNotImplementedFn("not"),
	"omit":                   makeNotImplementedFn("omit"),
	"or":                     makeNotImplementedFn("or"),
	"pick":                   makeNotImplementedFn("pick"),
	"print":                  makeNotImplementedFn("print"),
	"printf":                 makeNotImplementedFn("printf"),
	"println":                makeNotImplementedFn("println"),
	"randAlpha":              makeNotImplementedFn("randAlpha"),
	"randAlphaNum":           makeNotImplementedFn("randAlphaNum"),
	"randAscii":              makeNotImplementedFn("randAscii"),
	"randNumeric":            makeNotImplementedFn("randNumeric"),
	"regexFind":              makeNotImplementedFn("regexFind"),
	"regexMatch":             makeNotImplementedFn("regexMatch"),
	"regexReplaceAllLiteral": makeNotImplementedFn("regexReplaceAllLiteral"),
	"replace":                makeNotImplementedFn("replace"),
	"reverse":                makeNotImplementedFn("reverse"),
	"semver":                 makeNotImplementedFn("semver"),
	"semverCompare":          makeNotImplementedFn("semverCompare"),
	"set":                    makeNotImplementedFn("set"),
	"sha1sum":                makeNotImplementedFn("sha1sum"),
	"sha256sum":              makeNotImplementedFn("sha256sum"),
	"shuffle":                makeNotImplementedFn("shuffle"),
	"slice":                  makeNotImplementedFn("slice"),
	"split":                  makeNotImplementedFn("split"),
	"splitList":              makeNotImplementedFn("splitList"),
	"substr":                 makeNotImplementedFn("substr"),
	"ternary":                makeNotImplementedFn("ternary"),
	"toString":               makeNotImplementedFn("toString"),
	"trim":                   makeNotImplementedFn("trim"),
	"trimSuffix":             makeNotImplementedFn("trimSuffix"),
	"trunc":                  noopFn,
	"typeIs":                 noopFn,
	"typeIsLike":             noopFn,
	"typeOf":                 makeNotImplementedFn("typeOf"),
	"uniq":                   makeNotImplementedFn("uniq"),
	"urlquery":               makeNotImplementedFn("urlquery"),

	// Not implemented (surface usage in Strict tests)
	"fail":          makeNotImplementedFn("fail"),
	"lookup":        makeNotImplementedFn("lookup"),
	"getHostByName": makeNotImplementedFn("getHostByName"),

	// Frequently used sprig helper; treat as pass-through to avoid parse/exec failure
	"coalesce": passthrough1Fn,
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

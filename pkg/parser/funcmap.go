package parser

import "log/slog"

// tmplFunc is the signature for all analyzer template functions.
// Functions receive a context and Call (with unevaluated args), then return evalResult.
type tmplFunc func(ctx *evalCtx, call Call) evalResult

// funcMap holds real analyzer handlers used during evaluation.
var funcMap map[string]tmplFunc

// stubFuncMap lists all known function names for the parser (stubs only).
var stubFuncMap map[string]interface{}

// stub parser function for the stubFuncMap
var parseStub = func(...interface{}) interface{} { return nil }

func init() {
	// 1) Build evaluation registry with concrete handlers or not-implemented stubs.
	funcMap = map[string]tmplFunc{
		// Analysis-aware real handlers
		"include": includeFn,
		"tpl":     tplFn,
		"default": defaultFn,
		"get":     getFn,
		"index":   indexFn,

		// Pass-through helpers (preserve .Values usage)
		"indent":   emitArgsNoResultFn,
		"lower":    unaryPassThroughFn,
		"nindent":  emitArgsNoResultFn,
		"quote":    unaryPassThroughFn,
		"required": emitArgsNoResultFn,
		"upper":    unaryPassThroughFn,

		// Serializers (treated as unary pass-through for analysis)
		"toYaml":       unaryPassThroughFn,
		"mustToYaml":   unaryPassThroughFn,
		"toYamlPretty": unaryPassThroughFn,
		"toJson":       unaryPassThroughFn,
		"mustToJson":   unaryPassThroughFn,
		"toToml":       unaryPassThroughFn,

		// Data conversion: ignored in analysis
		"fromJson":      noopStrings,
		"fromJsonArray": noopStrings,
		"fromYaml":      noopStrings,
		"fromYamlArray": noopStrings,

		// Common collection/string helpers (not implemented yet)
		"and":                    makeNotImplementedFn("and"),
		"append":                 makeNotImplementedFn("append"),
		"b64dec":                 makeNotImplementedFn("b64dec"),
		"b64enc":                 makeNotImplementedFn("b64enc"),
		"call":                   makeNotImplementedFn("call"),
		"cat":                    makeNotImplementedFn("cat"),
		"concat":                 makeNotImplementedFn("concat"),
		"contains":               binaryEvalFn,
		"derivePassword":         makeNotImplementedFn("derivePassword"),
		"dict":                   dictFn,
		"empty":                  makeNotImplementedFn("empty"),
		"eq":                     binaryEvalFn,
		"first":                  makeNotImplementedFn("first"),
		"ge":                     binaryEvalFn,
		"gt":                     binaryEvalFn,
		"has":                    binaryEvalFn,
		"hasKey":                 binaryEvalFn,
		"hasPrefix":              binaryEvalFn,
		"hasSuffix":              binaryEvalFn,
		"html":                   makeNotImplementedFn("html"),
		"int":                    makeNotImplementedFn("int"),
		"join":                   makeNotImplementedFn("join"),
		"js":                     makeNotImplementedFn("js"),
		"keys":                   makeNotImplementedFn("keys"),
		"kindIs":                 binaryEvalFn,
		"le":                     binaryEvalFn,
		"len":                    makeNotImplementedFn("len"),
		"list":                   makeNotImplementedFn("list"),
		"lt":                     binaryEvalFn,
		"merge":                  makeNotImplementedFn("merge"),
		"mergeOverwrite":         makeNotImplementedFn("mergeOverwrite"),
		"ne":                     binaryEvalFn,
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
		"regexFind":              binaryEvalFn,
		"regexMatch":             binaryEvalFn,
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
		"trimPrefix":             binaryEvalFn,
		"trimSuffix":             binaryEvalFn,
		"trunc":                  noopStrings,
		"typeIs":                 binaryEvalFn,
		"typeIsLike":             binaryEvalFn,
		"typeOf":                 makeNotImplementedFn("typeOf"),
		"uniq":                   makeNotImplementedFn("uniq"),
		"urlquery":               makeNotImplementedFn("urlquery"),

		// Not implemented (surface usage in Strict tests)
		"fail":          makeNotImplementedFn("fail"),
		"lookup":        makeNotImplementedFn("lookup"),
		"getHostByName": makeNotImplementedFn("getHostByName"),

		// Frequently used sprig helper; emit args usage, no result folding
		"coalesce": emitArgsNoResultFn,
	}

	// 2) Build parse stubs from eval keys
	stubFuncMap = make(map[string]interface{}, len(funcMap))
	for name := range funcMap {
		stubFuncMap[name] = parseStub
	}
}

func getTemplateFunction(name string) tmplFunc {
	if fn, ok := funcMap[name]; ok {
		return fn
	}
	slog.Warn("unknown template function name", "name", name)
	return makeNotImplementedFn(name)
}

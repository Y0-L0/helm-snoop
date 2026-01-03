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
	// Build evaluation registry with concrete handlers.
	// Tests verify completeness against Sprig + Helm functions.
	funcMap = map[string]tmplFunc{
		// =================================================================
		// HELM-SPECIFIC FUNCTIONS
		// =================================================================
		"include":       includeFn, // Analysis-aware: resolves template definitions
		"tpl":           tplFn,     // Not implemented yet
		"toToml":        unarySerializeFn,
		"fromToml":      noopStrings,
		"toYaml":        unarySerializeFn,
		"mustToYaml":    unarySerializeFn,
		"toYamlPretty":  unarySerializeFn,
		"fromYaml":      noopStrings,
		"fromYamlArray": noopStrings,
		"toJson":        unarySerializeFn,
		"mustToJson":    unarySerializeFn,
		"fromJson":      noopStrings,
		"fromJsonArray": noopStrings,
		"lookup":        noopStrings,
		"fail":          emitArgsNoResultFn,
		"getHostByName": noopStrings,

		// =================================================================
		// GO TEMPLATE BUILT-INS (appear as functions in templates)
		// =================================================================
		"index":   indexFn, // Navigates paths by index/key
		"print":   emitArgsNoResultFn,
		"printf":  emitArgsNoResultFn,
		"println": emitArgsNoResultFn,
		"eq":      binaryEvalFn,
		"ne":      binaryEvalFn,
		"lt":      binaryEvalFn,
		"le":      binaryEvalFn,
		"gt":      binaryEvalFn,
		"ge":      binaryEvalFn,

		// =================================================================
		// SPRIG: ANALYSIS-AWARE IMPLEMENTATIONS
		// =================================================================
		"default": defaultFn, // Unions all argument paths
		"dict":    dictFn,    // Tracks key-value structure
		"get":     getFn,     // Navigates paths by key

		// =================================================================
		// SPRIG: COMPARISON/TEST FUNCTIONS
		// =================================================================
		"contains":   binaryEvalFn,
		"deepEqual":  binaryEvalFn,
		"has":        binaryEvalFn,
		"hasKey":     binaryEvalFn,
		"hasPrefix":  binaryEvalFn,
		"hasSuffix":  binaryEvalFn,
		"kindIs":     binaryEvalFn,
		"kindOf":     unaryPassThroughFn,
		"typeIs":     binaryEvalFn,
		"typeIsLike": binaryEvalFn,

		// =================================================================
		// SPRIG: LOGIC FUNCTIONS
		// =================================================================
		"all": emitArgsNoResultFn,
		"and": emitArgsNoResultFn,
		"any": emitArgsNoResultFn,
		"not": emitArgsNoResultFn,
		"or":  emitArgsNoResultFn,

		// =================================================================
		// SPRIG: STRING FUNCTIONS
		// =================================================================
		"abbrev":     unaryPassThroughFn,
		"abbrevboth": unaryPassThroughFn,
		"camelcase":  unaryPassThroughFn,
		"cat":        emitArgsNoResultFn,
		"indent":     emitArgsNoResultFn,
		"initials":   unaryPassThroughFn,
		"kebabcase":  unaryPassThroughFn,
		"lower":      unaryPassThroughFn,
		"nindent":    emitArgsNoResultFn,
		"nospace":    unaryPassThroughFn,
		"plural":     emitArgsNoResultFn,
		"quote":      unaryPassThroughFn,
		"repeat":     emitArgsNoResultFn,
		"replace":    emitArgsNoResultFn,
		"snakecase":  unaryPassThroughFn,
		"squote":     unaryPassThroughFn,
		"substr":     emitArgsNoResultFn,
		"swapcase":   unaryPassThroughFn,
		"title":      unaryPassThroughFn,
		"trim":       unaryPassThroughFn,
		"trimAll":    emitArgsNoResultFn,
		"trimPrefix": binaryEvalFn,
		"trimSuffix": binaryEvalFn,
		"trimall":    emitArgsNoResultFn,
		"trunc":      noopStrings,
		"untitle":    unaryPassThroughFn,
		"upper":      unaryPassThroughFn,
		"wrap":       emitArgsNoResultFn,
		"wrapWith":   emitArgsNoResultFn,

		// =================================================================
		// SPRIG: STRING LIST FUNCTIONS
		// =================================================================
		"join":      emitArgsNoResultFn,
		"sortAlpha": unaryPassThroughFn,
		"split":     emitArgsNoResultFn,
		"splitList": emitArgsNoResultFn,
		"splitn":    emitArgsNoResultFn,
		"toString":  unaryPassThroughFn,
		"toStrings": unaryPassThroughFn,

		// =================================================================
		// SPRIG: TYPE CONVERSION
		// =================================================================
		"atoi":      emitArgsNoResultFn,
		"float64":   emitArgsNoResultFn,
		"int":       emitArgsNoResultFn,
		"int64":     emitArgsNoResultFn,
		"toDecimal": unaryPassThroughFn,

		// =================================================================
		// SPRIG: MATH FUNCTIONS
		// =================================================================
		"add":     emitArgsNoResultFn,
		"add1":    emitArgsNoResultFn,
		"add1f":   emitArgsNoResultFn,
		"addf":    emitArgsNoResultFn,
		"biggest": emitArgsNoResultFn,
		"ceil":    emitArgsNoResultFn,
		"div":     emitArgsNoResultFn,
		"divf":    emitArgsNoResultFn,
		"floor":   emitArgsNoResultFn,
		"max":     emitArgsNoResultFn,
		"maxf":    emitArgsNoResultFn,
		"min":     emitArgsNoResultFn,
		"minf":    emitArgsNoResultFn,
		"mod":     emitArgsNoResultFn,
		"mul":     emitArgsNoResultFn,
		"mulf":    emitArgsNoResultFn,
		"round":   emitArgsNoResultFn,
		"sub":     emitArgsNoResultFn,
		"subf":    emitArgsNoResultFn,

		// =================================================================
		// SPRIG: DATE FUNCTIONS (no .Values tracking)
		// =================================================================
		"ago":              noopStrings,
		"date":             noopStrings,
		"dateInZone":       noopStrings,
		"dateModify":       noopStrings,
		"date_in_zone":     noopStrings,
		"date_modify":      noopStrings,
		"duration":         noopStrings,
		"durationRound":    noopStrings,
		"htmlDate":         noopStrings,
		"htmlDateInZone":   noopStrings,
		"mustDateModify":   noopStrings,
		"mustToDate":       noopStrings,
		"must_date_modify": noopStrings,
		"now":              noopStrings,
		"toDate":           noopStrings,
		"unixEpoch":        noopStrings,

		// =================================================================
		// SPRIG: DEFAULTS AND CONDITIONALS
		// =================================================================
		"coalesce": emitArgsNoResultFn,
		"empty":    emitArgsNoResultFn,
		"required": emitArgsNoResultFn,
		"ternary":  emitArgsNoResultFn,

		// =================================================================
		// SPRIG: LIST FUNCTIONS
		// =================================================================
		"append":      emitArgsNoResultFn,
		"chunk":       unaryPassThroughFn,
		"compact":     unaryPassThroughFn,
		"concat":      emitArgsNoResultFn,
		"first":       unaryPassThroughFn,
		"initial":     unaryPassThroughFn,
		"last":        unaryPassThroughFn,
		"list":        emitArgsNoResultFn,
		"mustAppend":  emitArgsNoResultFn,
		"mustChunk":   unaryPassThroughFn,
		"mustCompact": unaryPassThroughFn,
		"mustFirst":   unaryPassThroughFn,
		"mustHas":     binaryEvalFn,
		"mustInitial": unaryPassThroughFn,
		"mustLast":    unaryPassThroughFn,
		"mustPrepend": emitArgsNoResultFn,
		"mustPush":    emitArgsNoResultFn,
		"mustReverse": unaryPassThroughFn,
		"mustRest":    unaryPassThroughFn,
		"mustSlice":   unaryPassThroughFn,
		"mustUniq":    unaryPassThroughFn,
		"mustWithout": emitArgsNoResultFn,
		"prepend":     emitArgsNoResultFn,
		"push":        emitArgsNoResultFn,
		"rest":        unaryPassThroughFn,
		"reverse":     unaryPassThroughFn,
		"seq":         noopStrings,
		"shuffle":     unaryPassThroughFn,
		"slice":       emitArgsNoResultFn,
		"uniq":        unaryPassThroughFn,
		"until":       noopStrings,
		"untilStep":   noopStrings,
		"without":     emitArgsNoResultFn,

		// =================================================================
		// SPRIG: DICT FUNCTIONS
		// =================================================================
		"deepCopy":           unaryPassThroughFn,
		"dig":                emitArgsNoResultFn,
		"keys":               unaryPassThroughFn,
		"merge":              emitArgsNoResultFn,
		"mergeOverwrite":     emitArgsNoResultFn,
		"mustDeepCopy":       unaryPassThroughFn,
		"mustFromJson":       noopStrings,
		"mustMerge":          emitArgsNoResultFn,
		"mustMergeOverwrite": emitArgsNoResultFn,
		"omit":               omitPickFn,
		"pick":               omitPickFn,
		"pluck":              emitArgsNoResultFn,
		"set":                emitArgsNoResultFn,
		"unset":              emitArgsNoResultFn,
		"values":             unaryPassThroughFn,

		// =================================================================
		// SPRIG: ENCODING FUNCTIONS
		// =================================================================
		"b32dec": unaryPassThroughFn,
		"b32enc": unaryPassThroughFn,
		"b64dec": unaryPassThroughFn,
		"b64enc": unaryPassThroughFn,

		// =================================================================
		// SPRIG: CRYPTO FUNCTIONS (no .Values tracking)
		// =================================================================
		"adler32sum":               noopStrings,
		"bcrypt":                   emitArgsNoResultFn,
		"buildCustomCert":          noopStrings,
		"decryptAES":               emitArgsNoResultFn,
		"derivePassword":           emitArgsNoResultFn,
		"encryptAES":               emitArgsNoResultFn,
		"genCA":                    noopStrings,
		"genCAWithKey":             noopStrings,
		"genPrivateKey":            noopStrings,
		"genSelfSignedCert":        noopStrings,
		"genSelfSignedCertWithKey": noopStrings,
		"genSignedCert":            noopStrings,
		"genSignedCertWithKey":     noopStrings,
		"htpasswd":                 emitArgsNoResultFn,
		"randAlpha":                noopStrings,
		"randAlphaNum":             noopStrings,
		"randAscii":                noopStrings,
		"randBytes":                noopStrings,
		"randInt":                  noopStrings,
		"randNumeric":              noopStrings,
		"sha1sum":                  emitArgsNoResultFn,
		"sha256sum":                emitArgsNoResultFn,
		"sha512sum":                emitArgsNoResultFn,
		"uuidv4":                   noopStrings,

		// =================================================================
		// SPRIG: REGEX FUNCTIONS
		// =================================================================
		"mustRegexFind":              binaryEvalFn,
		"mustRegexFindAll":           emitArgsNoResultFn,
		"mustRegexMatch":             binaryEvalFn,
		"mustRegexReplaceAll":        emitArgsNoResultFn,
		"mustRegexReplaceAllLiteral": emitArgsNoResultFn,
		"mustRegexSplit":             emitArgsNoResultFn,
		"regexFind":                  binaryEvalFn,
		"regexFindAll":               emitArgsNoResultFn,
		"regexMatch":                 binaryEvalFn,
		"regexQuoteMeta":             unaryPassThroughFn,
		"regexReplaceAll":            emitArgsNoResultFn,
		"regexReplaceAllLiteral":     emitArgsNoResultFn,
		"regexSplit":                 emitArgsNoResultFn,

		// =================================================================
		// SPRIG: PATH/URL FUNCTIONS
		// =================================================================
		"base":     unaryPassThroughFn,
		"clean":    unaryPassThroughFn,
		"dir":      unaryPassThroughFn,
		"ext":      unaryPassThroughFn,
		"isAbs":    emitArgsNoResultFn,
		"osBase":   unaryPassThroughFn,
		"osClean":  unaryPassThroughFn,
		"osDir":    unaryPassThroughFn,
		"osExt":    unaryPassThroughFn,
		"osIsAbs":  emitArgsNoResultFn,
		"urlJoin":  emitArgsNoResultFn,
		"urlParse": emitArgsNoResultFn,
		"urlquery": unaryPassThroughFn,

		// =================================================================
		// SPRIG: SEMANTIC VERSION FUNCTIONS
		// =================================================================
		"semver":        emitArgsNoResultFn,
		"semverCompare": binaryEvalFn,

		// =================================================================
		// SPRIG: REFLECTION/TYPE FUNCTIONS
		// =================================================================
		"hello":  noopStrings,
		"len":    emitArgsNoResultFn,
		"tuple":  emitArgsNoResultFn,
		"typeOf": unaryPassThroughFn,

		// =================================================================
		// SPRIG: TEMPLATE FUNCTIONS
		// =================================================================
		"call": emitArgsNoResultFn,
		"html": unaryPassThroughFn,
		"js":   unaryPassThroughFn,

		// =================================================================
		// SPRIG: JSON/YAML FUNCTIONS (with must variants)
		// =================================================================
		"mustToPrettyJson": noopStrings,
		"mustToRawJson":    noopStrings,
		"toPrettyJson":     unarySerializeFn,
		"toRawJson":        unarySerializeFn,
	}

	// Build parse stubs from eval keys
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

package parser

import (
	"fmt"
	"log/slog"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

func includeFn(ctx *evalCtx, call Call) evalResult {
	// 1. Validate arguments - include requires at least 1 argument (template name)
	if len(call.Args) < 1 {
		slog.Warn("include: requires at least 1 argument", "count", len(call.Args))
		Must("include: requires at least 1 argument")
		return evalResult{}
	}

	// 2. Extract template name from first argument
	nameResult := ctx.Eval(call.Args[0])
	ctx.Emit(nameResult.paths...)

	// 3. Determine context for template body
	var templatePrefix *path.Path
	var dictPaths map[string]*path.Path
	var dictLits map[string]string
	isRootContext := false

	if len(call.Args) >= 2 {
		ctxArg := call.Args[1]

		if varNode, ok := ctxArg.(*parse.VariableNode); ok && len(varNode.Ident) > 0 && varNode.Ident[0] == "$" {
			isRootContext = true
		} else {
			ctxResult := ctx.Eval(ctxArg)

			if ctxResult.dict != nil || ctxResult.dictLits != nil {
				dictPaths = ctxResult.dict
				dictLits = ctxResult.dictLits
			} else {
				ctx.Emit(ctxResult.paths...)
				if len(ctxResult.paths) > 0 {
					// Only set prefix for non-empty paths
					// Empty path / means root context without dict params,
					// so relative fields are unresolved parameter names, not .Values paths
					if ctxResult.paths[0].ID() != "/" {
						templatePrefix = ctxResult.paths[0]
					}
				}
			}
		}
	}

	// Extract template name from literal strings
	if len(nameResult.args) == 0 {
		slog.Warn("include: template name must be a string literal")
		Must("include: template name must be a string literal")
		return evalResult{}
	}
	templateName := nameResult.args[0]

	// 4. Check template index availability
	if ctx.idx == nil {
		slog.Warn("include: template index not available")
		Must("include: template index not available")
		return evalResult{}
	}

	// 5. Look up the template definition
	tmplDef, found := ctx.idx.get(templateName)
	if !found {
		slog.Warn("include: template not found", "name", templateName)
		Must(fmt.Sprintf("include: template %q not found", templateName))
		return evalResult{}
	}

	// 6. Check for circular includes using inStack
	if ctx.inStack[templateName] {
		panic(fmt.Sprintf("include: circular dependency on template %q", templateName))
	}

	// 7. Mark template as in-stack before evaluation
	ctx.inStack[templateName] = true
	defer func() {
		// Clean up stack after evaluation (even if panic occurs)
		delete(ctx.inStack, templateName)
	}()

	slog.Debug("include: evaluating template body", "name", templateName, "file", tmplDef.file)

	// 8. Evaluate template body with context
	var restore func()
	if isRootContext {
		restore = ctx.WithPrefixes(nil)
	} else if dictPaths != nil || dictLits != nil {
		restore = ctx.WithDictParams(dictPaths, dictLits)
	} else if templatePrefix != nil {
		restore = ctx.WithPrefixes(path.Paths{templatePrefix})
	} else {
		restore = func() {}
	}

	ctx.Eval(tmplDef.root)
	restore()

	// Include doesn't return a value for path analysis
	// (paths are emitted during evaluation of the template body)
	return evalResult{}
}

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

// makeNotImplementedFn logs and fails in strict mode.
func makeNotImplementedFn(name string) tmplFunc {
	return func(_ *evalCtx, _ Call) evalResult {
		slog.Warn("template function not implemented", "name", name)
		Must("template function not implemented: " + name)
		return evalResult{}
	}
}

func noopStrings(_ *evalCtx, _ Call) evalResult {
	return evalResult{}
}

// ==============================================================================
// FLAVOR 1: COMPLEX VALUE PRODUCERS
// ==============================================================================

// dictFn builds both conservative union AND structure tracking
func dictFn(ctx *evalCtx, call Call) evalResult {
	dict := make(map[string]*path.Path)
	dictLits := make(map[string]string)
	var allPaths []*path.Path

	for i := 0; i+1 < len(call.Args); i += 2 {
		keyResult := ctx.Eval(call.Args[i])
		valResult := ctx.Eval(call.Args[i+1])

		allPaths = append(allPaths, valResult.paths...)

		if len(keyResult.args) != 1 {
			continue
		}
		key := keyResult.args[0]

		if len(valResult.paths) > 0 {
			dict[key] = valResult.paths[0]
		}

		if len(valResult.args) > 0 {
			dictLits[key] = valResult.args[0]
		}
	}

	return evalResult{
		paths:    allPaths,
		dict:     dict,
		dictLits: dictLits,
	}
}

// indexFn builds child paths by appending keys
func indexFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) < 2 {
		slog.Warn("index: need at least 2 args", "count", len(call.Args))
		Must("index: need at least 2 args")
		return evalResult{}
	}

	baseResult := ctx.Eval(call.Args[0])

	// Check for dict structure
	if baseResult.dict != nil {
		var resolvedPaths []*path.Path

		for _, arg := range call.Args[1:] {
			keyResult := ctx.Eval(arg)
			for _, key := range keyResult.args {
				if p, ok := baseResult.dict[key]; ok {
					resolvedPaths = append(resolvedPaths, p)
				}
			}
		}

		return evalResult{paths: resolvedPaths}
	}

	// Fall back to conservative: append keys to all paths
	var keys []string
	for _, arg := range call.Args[1:] {
		keyResult := ctx.Eval(arg)
		keys = append(keys, keyResult.args...)
	}

	var modifiedPaths []*path.Path
	for _, base := range baseResult.paths {
		p := *base
		for _, key := range keys {
			p = p.WithAny(key)
		}
		modifiedPaths = append(modifiedPaths, &p)
	}

	return evalResult{paths: modifiedPaths}
}

// getFn is like index but with exactly 2 args
func getFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) != 2 {
		slog.Warn("get: need exactly 2 args", "count", len(call.Args))
		Must("get: need exactly 2 args")
		return evalResult{}
	}

	baseResult := ctx.Eval(call.Args[0])
	keyResult := ctx.Eval(call.Args[1])

	// Check for dict structure
	if baseResult.dict != nil && len(keyResult.args) == 1 {
		if p, ok := baseResult.dict[keyResult.args[0]]; ok {
			return evalResult{paths: []*path.Path{p}}
		}
		return evalResult{}
	}

	// Fall back to conservative
	if len(keyResult.args) != 1 {
		slog.Warn("get: key must be exactly one literal", "count", len(keyResult.args))
		Must("get: key must be exactly one literal")
		return evalResult{}
	}

	var modifiedPaths []*path.Path
	for _, base := range baseResult.paths {
		p := *base
		p = p.WithAny(keyResult.args[0])
		modifiedPaths = append(modifiedPaths, &p)
	}

	return evalResult{paths: modifiedPaths}
}

// defaultFn unions all argument paths and returns them.
func defaultFn(ctx *evalCtx, call Call) evalResult {
	var allPaths path.Paths
	for _, arg := range call.Args {
		result := ctx.Eval(arg)
		allPaths = append(allPaths, result.paths...)
	}

	return evalResult{paths: allPaths}
}

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

	// 3. Determine prefix for template body based on context argument
	// The context argument (if present) becomes "." inside the template
	var templatePrefix *path.Path
	isRootContext := false

	if len(call.Args) >= 2 {
		ctxArg := call.Args[1]

		// Check if context argument is $ (root variable)
		if varNode, ok := ctxArg.(*parse.VariableNode); ok && len(varNode.Ident) > 0 && varNode.Ident[0] == "$" {
			// $ refers to root context -> no prefix for template body
			isRootContext = true
		} else {
			// Evaluate the context argument to determine prefix
			ctxResult := ctx.Eval(ctxArg)
			ctx.Emit(ctxResult.paths...)

			// Only set prefix for simple contexts, not dicts
			// dict != nil indicates the context is a dict structure (from dictFn)
			// Dict context handling is not yet supported - would need to map dict keys to paths
			// For now, we skip prefix setting to avoid false positives like /foo/value, /foo/scope
			if ctxResult.dict == nil && len(ctxResult.paths) > 0 {
				templatePrefix = ctxResult.paths[0]
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

	// 8. Evaluate the template body with the appropriate prefix
	// The context argument to include becomes "." inside the template
	// - If arg is $: no prefix (root context)
	// - If arg is .: current prefix
	// - If arg is .Values.foo: prefix = foo
	var restore func()
	if isRootContext {
		// $ argument -> clear prefix for template body
		restore = ctx.WithPrefix(nil)
	} else if templatePrefix != nil {
		// Explicit context -> use it as prefix
		restore = ctx.WithPrefix(templatePrefix)
	} else {
		// No context arg or DotNode -> keep current prefix
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
	// Always create a non-nil map, even if empty
	// This serves as a marker that the result came from a dict call
	// (nil dict = not from dict, non-nil dict = from dict)
	dict := make(map[string]*path.Path)
	var allPaths []*path.Path

	// Parse key1, val1, key2, val2, ...
	for i := 0; i+1 < len(call.Args); i += 2 {
		keyResult := ctx.Eval(call.Args[i])
		valResult := ctx.Eval(call.Args[i+1])

		// Conservative union
		allPaths = append(allPaths, valResult.paths...)

		// Structure tracking - populate map when we can resolve keys
		if len(keyResult.args) == 1 && len(valResult.paths) > 0 {
			dict[keyResult.args[0]] = valResult.paths[0]
		}
	}

	return evalResult{
		paths: allPaths,
		dict:  dict, // Non-nil map (possibly empty) marks this as dict result
	}
}

// indexFn builds child paths by appending keys
func indexFn(ctx *evalCtx, call Call) evalResult {
	if len(call.Args) < 2 {
		slog.Warn("index: need at least 2 args", "count", len(call.Args))
		Must("index: need at least 2 args")
		return evalResult{}
	}

	// Eval base arg
	baseResult := ctx.Eval(call.Args[0])

	// Check for dict structure
	if baseResult.dict != nil {
		var resolvedPaths []*path.Path

		// Extract keys from remaining args
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

// defaultFn unions all argument paths
func defaultFn(ctx *evalCtx, call Call) evalResult {
	var allPaths []*path.Path
	for _, arg := range call.Args {
		result := ctx.Eval(arg)
		allPaths = append(allPaths, result.paths...)
	}

	return evalResult{paths: allPaths}
}

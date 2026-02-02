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
	ctx.Emit(call.Node.Position(), nameResult.paths...)

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
				ctx.Emit(call.Node.Position(), ctxResult.paths...)
				if len(ctxResult.paths) > 0 {
					// Only set prefix for non-empty paths
					// Empty path means root context without dict params,
					// so relative fields are unresolved parameter names, not .Values paths
					if ctxResult.paths[0].ID() != "." {
						templatePrefix = ctxResult.paths[0]
					}
				}
			}
		}
	}

	// Extract template name from literal strings
	if len(nameResult.args) == 0 {
		slog.Info("include: template name must be a string literal")
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

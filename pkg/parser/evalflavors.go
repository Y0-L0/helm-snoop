package parser

import (
	"log/slog"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// isBuiltinObject checks if a field name is a Helm built-in object
func isBuiltinObject(field string) bool {
	return field == "Release" || field == "Chart" ||
		field == "Files" || field == "Capabilities" ||
		field == "Template"
}

func (e *evalCtx) evalParamPaths(firstField string, restFields []string) (evalResult, bool) {
	if e.paramPaths == nil {
		return evalResult{}, false
	}
	basePath, ok := e.paramPaths[firstField]
	if !ok {
		return evalResult{}, false
	}

	if len(restFields) > 0 && isBuiltinObject(restFields[0]) {
		return evalResult{}, true
	}

	p := *basePath
	for _, field := range restFields {
		p = p.WithKey(field)
	}
	return evalResult{paths: []*path.Path{&p}}, true
}

func (e *evalCtx) evalParamLits(firstField string, restFields []string) (evalResult, bool) {
	if e.paramLits == nil {
		return evalResult{}, false
	}
	litVal, ok := e.paramLits[firstField]
	if !ok {
		return evalResult{}, false
	}
	if len(restFields) > 0 {
		slog.Warn("field access on literal parameter", "param", firstField, "literal", litVal)
		return evalResult{}, true
	}
	slog.Debug("resolved literal parameter", "param", firstField, "value", litVal)

	return evalResult{args: []string{litVal}}, true
}

// evalFieldNode evaluates field access like .Values.foo.bar
func (e *evalCtx) evalFieldNode(node *parse.FieldNode) evalResult {
	if len(node.Ident) == 0 {
		return evalResult{}
	}

	firstField := node.Ident[0]
	restFields := node.Ident[1:]

	if result, ok := e.evalParamPaths(firstField, restFields); ok {
		return result
	}

	if result, ok := e.evalParamLits(firstField, restFields); ok {
		return result
	}

	if firstField == "Values" {
		if len(node.Ident) == 1 {
			// Just ".Values" with no path
			return evalResult{}
		}
		p := path.NewPath(restFields...)
		return evalResult{paths: []*path.Path{p}}
	}

	if isBuiltinObject(firstField) {
		return evalResult{}
	}

	// Only track relative field access inside with/range blocks
	// (outside with/range, relative fields like .foo are not .Values)
	if !e.hasPrefix() {
		return evalResult{}
	}

	p := path.NewPath(node.Ident...)
	prefixed := e.addPrefixes(p)
	return evalResult{paths: prefixed}
}

// evalChainNode evaluates (.foo).bar.baz
func (e *evalCtx) evalChainNode(node *parse.ChainNode) evalResult {
	if node.Node == nil {
		return evalResult{}
	}

	baseResult := e.Eval(node.Node)

	if len(node.Field) == 0 {
		return baseResult
	}

	var modifiedPaths []*path.Path
	for _, basePath := range baseResult.paths {
		p := *basePath
		for _, field := range node.Field {
			p = p.WithKey(field)
		}
		modifiedPaths = append(modifiedPaths, &p)
	}

	return evalResult{
		paths: modifiedPaths,
		args:  baseResult.args,
	}
}

// evalCommandNode evaluates a command (either a function call or single field/literal)
func (e *evalCtx) evalCommandNode(node *parse.CommandNode) evalResult {
	if len(node.Args) == 0 {
		return evalResult{}
	}

	// Check if first arg is an identifier (function name)
	id, ok := node.Args[0].(*parse.IdentifierNode)
	if !ok {
		// Not a function call, just a single field/literal
		if len(node.Args) == 1 {
			return e.Eval(node.Args[0])
		}
		slog.Warn("command with multiple non-function args", "count", len(node.Args))
		return evalResult{}
	}

	// Build Call with unevaluated args
	call := Call{
		Name: id.Ident,
		Args: node.Args[1:], // Remaining args (unevaluated)
		Node: node,
	}

	// Get and invoke the function
	fn := getTemplateFunction(id.Ident)
	return fn(e, call)
}

// evalPipeNode evaluates a pipeline recursively from the outside in.
// For {{ .Values.foo | quote | upper }}, we evaluate upper with (foo | quote) as an unevaluated arg.
func (e *evalCtx) evalPipeNode(node *parse.PipeNode) evalResult {
	if len(node.Cmds) == 0 {
		return evalResult{}
	}

	// Base case: single command (no pipe)
	if len(node.Cmds) == 1 {
		return e.evalCommandNode(node.Cmds[0])
	}

	// Recursive case: pipe the left side into the right side
	// For [A, B, C], we want C to receive (A | B) as its last argument

	lastCmd := node.Cmds[len(node.Cmds)-1]

	// Check if last command is a function call
	if len(lastCmd.Args) == 0 {
		return evalResult{}
	}

	id, ok := lastCmd.Args[0].(*parse.IdentifierNode)
	if !ok {
		// Not a function, just evaluate the whole pipe left-to-right as fallback
		var lastResult evalResult
		for _, cmd := range node.Cmds {
			lastResult = e.evalCommandNode(cmd)
		}
		// Emit paths if not already emitted
		e.Emit(node.Pos, lastResult.paths...)
		return lastResult
	}

	// Create synthetic PipeNode for everything before the last command
	var pipedArg parse.Node
	if len(node.Cmds) == 2 {
		// Just one command before, use it directly
		pipedArg = node.Cmds[0]
	} else {
		// Multiple commands before, create sub-pipe
		pipedArg = &parse.PipeNode{
			Cmds: node.Cmds[:len(node.Cmds)-1],
		}
	}

	// Build args list with piped value appended
	args := make([]parse.Node, len(lastCmd.Args)-1, len(lastCmd.Args))
	copy(args, lastCmd.Args[1:])
	args = append(args, pipedArg)

	// Build Call with normalized args
	call := Call{
		Name: id.Ident,
		Args: args,
		Node: lastCmd,
	}

	// Invoke function (it will recursively eval the piped arg)
	fn := getTemplateFunction(id.Ident)
	return fn(e, call)
}

// evalIfNode evaluates an if/else control flow node.
// Control flow nodes emit their condition paths directly (not wrapped in ActionNode).
func (e *evalCtx) evalIfNode(node *parse.IfNode) evalResult {
	// Evaluate condition and emit paths
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		e.Emit(node.Pipe.Pos, result.paths...)
	}

	// Evaluate "then" branch
	if node.List != nil {
		e.Eval(node.List)
	}

	// Evaluate "else" branch
	if node.ElseList != nil {
		e.Eval(node.ElseList)
	}

	return evalResult{}
}

// evalRangeNode evaluates a range loop control flow node.
// The range expression sets the context prefix but is not emitted itself.
// Paths are only emitted when actually accessed via . inside the body.
// If the range expression returns multiple paths (e.g., from concat),
// all paths are set as prefixes and field access generates paths for each.
// Variable declarations bind the value variable to the wildcard path.
func (e *evalCtx) evalRangeNode(node *parse.RangeNode) evalResult {
	var rangePrefixes path.Paths
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		for _, p := range result.paths {
			wildcardPath := p.WithWildcard()
			rangePrefixes = append(rangePrefixes, &wildcardPath)
		}
	}

	if node.List != nil {
		restore := e.WithVariables(node.Pipe, rangePrefixes, true)
		e.Eval(node.List)
		restore()
	}

	if node.ElseList != nil {
		e.Eval(node.ElseList)
	}

	return evalResult{}
}

// evalWithNode evaluates a with scoping control flow node.
// The with expression sets the context prefix but is not emitted itself.
// Paths are only emitted when actually accessed via . inside the body.
// If the with expression returns multiple paths (e.g., from concat or default),
// all paths are set as prefixes and field access generates paths for each.
// Variable declarations bind the variable to the path.
func (e *evalCtx) evalWithNode(node *parse.WithNode) evalResult {
	var withPrefixes path.Paths
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		withPrefixes = result.paths
	}

	if node.List != nil {
		restore := e.WithVariables(node.Pipe, withPrefixes, false)
		e.Eval(node.List)
		restore()
	}

	if node.ElseList != nil {
		e.Eval(node.ElseList)
	}

	return evalResult{}
}

// evalTemplateNode evaluates a template action node.
// Template actions like {{ template "name" pipeline }} evaluate the pipeline argument.
func (e *evalCtx) evalTemplateNode(node *parse.TemplateNode) evalResult {
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		if result.dict == nil && result.dictLits == nil {
			e.Emit(node.Pipe.Pos, result.paths...)
		}
	}
	return evalResult{}
}

// evalVariableNode handles $ root context and range/with variables.
func (e *evalCtx) evalVariableNode(node *parse.VariableNode) evalResult {
	if len(node.Ident) == 0 {
		return evalResult{}
	}

	firstIdent := node.Ident[0]
	if len(firstIdent) == 0 || firstIdent[0] != '$' {
		return evalResult{}
	}

	// Handle bare $ (root context)
	if firstIdent == "$" {
		if len(node.Ident) == 1 {
			return evalResult{paths: path.Paths{path.NewPath()}}
		}

		if node.Ident[1] == "Values" {
			if len(node.Ident) < 3 {
				return evalResult{}
			}
			p := path.NewPath(node.Ident[2:]...)
			return evalResult{paths: path.Paths{p}}
		}

		return evalResult{}
	}

	// Check for range/with variables: strip $ prefix
	varName := firstIdent[1:]

	if e.variables != nil {
		if basePath, ok := e.variables[varName]; ok {
			if len(node.Ident) == 1 {
				return evalResult{paths: []*path.Path{basePath}}
			}
			p := *basePath
			for _, field := range node.Ident[1:] {
				p = p.WithKey(field)
			}
			return evalResult{paths: []*path.Path{&p}}
		}
	}

	return evalResult{}
}

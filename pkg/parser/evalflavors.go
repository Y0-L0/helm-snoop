package parser

import (
	"log/slog"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// evalFieldNode evaluates field access like .Values.foo.bar
func (e *evalCtx) evalFieldNode(node *parse.FieldNode) evalResult {
	if len(node.Ident) == 0 {
		return evalResult{}
	}

	// Check if it starts with "Values"
	if node.Ident[0] == "Values" {
		if len(node.Ident) == 1 {
			// Just ".Values" with no path
			return evalResult{}
		}
		// Build absolute path from .Values.foo.bar -> ["foo", "bar"]
		p := path.NewPath(node.Ident[1:]...)
		return evalResult{paths: []*path.Path{p}}
	}

	// Only track relative field access inside with/range blocks
	// (outside with/range, relative fields like .Release.Name are not .Values)
	if !e.hasPrefix() {
		return evalResult{}
	}

	// Build path with prefix
	p := path.NewPath(node.Ident...)
	prefixed := e.addPrefix(p)
	return evalResult{paths: []*path.Path{prefixed}}
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
		e.Emit(lastResult.paths...)
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
		e.Emit(result.paths...)
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
// Control flow nodes emit their range expression paths directly (not wrapped in ActionNode).
func (e *evalCtx) evalRangeNode(node *parse.RangeNode) evalResult {
	// Evaluate range expression and emit paths
	var rangePrefix *path.Path
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		e.Emit(result.paths...)
		// Set prefix for range body: items -> items.*
		// The path from Eval() is already prefixed if needed
		if len(result.paths) > 0 {
			p := result.paths[0].WithKey("*")
			rangePrefix = &p
		}
	}

	// Evaluate range body with prefix
	if node.List != nil {
		restore := e.WithPrefix(rangePrefix)
		e.Eval(node.List)
		restore() // Restore original context before else branch
	}

	// Evaluate else branch if present (no prefix in else)
	if node.ElseList != nil {
		e.Eval(node.ElseList)
	}

	return evalResult{}
}

// evalWithNode evaluates a with scoping control flow node.
// Control flow nodes emit their with expression paths directly (not wrapped in ActionNode).
func (e *evalCtx) evalWithNode(node *parse.WithNode) evalResult {
	// Evaluate with expression and emit paths
	var withPrefix *path.Path
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		e.Emit(result.paths...)
		// Set prefix for with body
		// The path from Eval() is already prefixed if needed
		if len(result.paths) > 0 {
			withPrefix = result.paths[0]
		}
	}

	// Evaluate with body with prefix
	if node.List != nil {
		restore := e.WithPrefix(withPrefix)
		e.Eval(node.List)
		restore() // Restore original context before else branch
	}

	// Evaluate else branch if present (no prefix in else)
	if node.ElseList != nil {
		e.Eval(node.ElseList)
	}

	return evalResult{}
}

// evalTemplateNode evaluates a template action node.
// Template actions like {{ template "name" pipeline }} evaluate the pipeline argument.
func (e *evalCtx) evalTemplateNode(node *parse.TemplateNode) evalResult {
	// Evaluate the pipeline argument if present
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		// Emit pipeline paths
		e.Emit(result.paths...)
	}
	// Note: We don't evaluate the template body itself here
	// That would require template resolution like include
	return evalResult{}
}

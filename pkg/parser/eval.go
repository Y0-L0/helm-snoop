package parser

import (
	"log/slog"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// Call represents an invocation of a template function during analysis.
type Call struct {
	// Name is the function identifier (e.g., "index", "default", "toYaml").
	Name string

	// Args contains the unevaluated argument nodes.
	// Functions must call ctx.Eval() on each arg to analyze them.
	// When a function is called via pipe, the piped value is appended as the last arg.
	Args []parse.Node

	// Node is the original CommandNode for diagnostics and error reporting.
	Node parse.Node
}

type evalCtx struct {
	tree     *parse.Tree
	out      *path.Paths
	prefix   *path.Path
	idx      *TemplateIndex
	inStack  map[string]bool
	depth    int
	maxDepth int
}

// newEvalCtx constructs an evaluation context with initialized state.
func newEvalCtx(tree *parse.Tree, out *path.Paths, idx *TemplateIndex) *evalCtx {
	return &evalCtx{
		tree:     tree,
		out:      out,
		idx:      idx,
		inStack:  make(map[string]bool),
		depth:    0,
		maxDepth: 64,
	}
}

func (e *evalCtx) Emit(p *path.Path) {
	if p == nil || e == nil || e.out == nil {
		panic("Invalid state in Emit function")
	}
	// TODO: Add prefix support alongside evalCtx.WithPrefix(p *path.Path)
	e.out.Append(p)
}

// TODO: Implement WithPrefix behaviour for with and range blocks.
func (e *evalCtx) WithPrefix(p *path.Path) func() {
	return nil
}

// Eval recursively evaluates a node and returns paths, strings, and structure.
// This is the primary entry point for functions to analyze their arguments.
func (e *evalCtx) Eval(n parse.Node) evalResult {
	// Guard against infinite recursion
	e.depth++
	if e.depth > e.maxDepth {
		slog.Error("maximum recursion depth exceeded", "depth", e.depth)
		panic("maximum recursion depth exceeded")
	}
	defer func() { e.depth-- }()

	if n == nil {
		return evalResult{}
	}

	switch node := n.(type) {
	// Interesting cases - structural and control flow nodes
	case *parse.ListNode:
		// ListNode contains a list of nodes (the template root)
		// Eval each node in sequence
		var lastResult evalResult
		if node.Nodes != nil {
			for _, child := range node.Nodes {
				lastResult = e.Eval(child)
			}
		}
		return lastResult

	case *parse.ActionNode:
		// ActionNode wraps a pipeline ({{ ... }})
		if node.Pipe != nil {
			result := e.Eval(node.Pipe)
			// Emit paths at the top level (ActionNode), not at every pipe level
			// This prevents duplicate emissions from nested pipes
			for _, p := range result.paths {
				e.Emit(p)
			}
			return result
		}
		return evalResult{}

	case *parse.FieldNode:
		return e.evalFieldNode(node)

	case *parse.CommandNode:
		return e.evalCommandNode(node)

	case *parse.PipeNode:
		return e.evalPipeNode(node)

	case *parse.IfNode:
		return e.evalIfNode(node)

	case *parse.RangeNode:
		return e.evalRangeNode(node)

	case *parse.WithNode:
		return e.evalWithNode(node)

	case *parse.TemplateNode:
		return e.evalTemplateNode(node)

	// Literal value nodes
	case *parse.StringNode:
		return evalResult{args: []string{node.Text}}

	case *parse.NumberNode:
		return evalResult{args: []string{node.Text}}

	case *parse.BoolNode:
		// Return boolean as string
		if node.True {
			return evalResult{args: []string{"true"}}
		}
		return evalResult{args: []string{"false"}}

	// No-op cases - nodes that don't contribute to path analysis
	case *parse.TextNode:
		// Plain text outside of actions - no analysis needed
		return evalResult{}

	case *parse.CommentNode:
		// Template comments - no analysis needed
		return evalResult{}

	case *parse.DotNode:
		// "." refers to current context (not .Values)
		return evalResult{}

	case *parse.VariableNode:
		// Variables like $foo - not implemented yet
		return evalResult{}

	case *parse.IdentifierNode:
		// Bare identifier (shouldn't happen in well-formed templates)
		return evalResult{}

	case *parse.NilNode:
		// Nil constant
		return evalResult{}

	default:
		slog.Warn("unsupported node type in Eval", "type", node.Type())
		Must("unsupported node type in Eval")
		return evalResult{}
	}
}

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
		// Build path from .Values.foo.bar -> ["foo", "bar"]
		p := path.NewPath(node.Ident[1:]...)
		return evalResult{paths: []*path.Path{p}}
	}

	// Non-.Values fields (e.g., .Release, .Chart, context fields)
	// Return empty for now
	return evalResult{}
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
		for _, p := range lastResult.paths {
			e.Emit(p)
		}
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
		for _, p := range result.paths {
			e.Emit(p)
		}
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
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		for _, p := range result.paths {
			e.Emit(p)
		}
	}

	// Evaluate range body
	if node.List != nil {
		e.Eval(node.List)
	}

	// Evaluate else branch if present
	if node.ElseList != nil {
		e.Eval(node.ElseList)
	}

	return evalResult{}
}

// evalWithNode evaluates a with scoping control flow node.
// Control flow nodes emit their with expression paths directly (not wrapped in ActionNode).
func (e *evalCtx) evalWithNode(node *parse.WithNode) evalResult {
	// Evaluate with expression and emit paths
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		for _, p := range result.paths {
			e.Emit(p)
		}
	}

	// Evaluate with body
	if node.List != nil {
		e.Eval(node.List)
	}

	// Evaluate else branch if present
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
		for _, p := range result.paths {
			e.Emit(p)
		}
	}
	// Note: We don't evaluate the template body itself here
	// That would require template resolution like include
	return evalResult{}
}

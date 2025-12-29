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

func (e *evalCtx) Emit(paths ...*path.Path) {
	if e == nil || e.out == nil {
		panic("Invalid state in Emit function")
	}
	for _, p := range paths {
		if p == nil {
			panic("Invalid state in Emit function")
		}
		e.out.Append(p)
	}
}

// addPrefix returns a new path with the context prefix prepended if one is set.
// Used when creating relative paths inside with/range blocks.
// Returns the original path if no prefix is set.
func (e *evalCtx) addPrefix(p *path.Path) *path.Path {
	if !e.hasPrefix() {
		return p
	}
	prefixed := e.prefix.Join(*p)
	return &prefixed
}

// WithPrefix sets a context prefix and returns a cleanup function.
// Used by with/range to change the meaning of "." in nested scopes.
// If p is nil, no prefix is set (no-op that still returns a valid cleanup function).
func (e *evalCtx) WithPrefix(p *path.Path) func() {
	oldPrefix := e.prefix
	e.prefix = p
	return func() {
		e.prefix = oldPrefix
	}
}

// hasPrefix returns true if a context prefix is currently set.
func (e *evalCtx) hasPrefix() bool {
	return e.prefix != nil
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
	case *parse.ListNode:
		return e.evalListNode(node)
	case *parse.ActionNode:
		return e.evalActionNode(node)
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

	// No-op cases - nodes that don't contribute to path analysis
	case *parse.BoolNode:
		return evalResult{}
	case *parse.TextNode:
		return evalResult{}
	case *parse.CommentNode:
		return evalResult{}
	case *parse.DotNode:
		return evalResult{}
	case *parse.VariableNode:
		return evalResult{}
	case *parse.IdentifierNode:
		return evalResult{}
	case *parse.NilNode:
		return evalResult{}

	default:
		slog.Warn("unsupported node type in Eval", "type", node.Type())
		Must("unsupported node type in Eval")
		return evalResult{}
	}
}

// evalListNode evaluates a list of nodes in sequence.
func (e *evalCtx) evalListNode(node *parse.ListNode) evalResult {
	var lastResult evalResult
	if node.Nodes != nil {
		for _, child := range node.Nodes {
			lastResult = e.Eval(child)
		}
	}
	return lastResult
}

// evalActionNode evaluates an action node ({{ ... }}).
func (e *evalCtx) evalActionNode(node *parse.ActionNode) evalResult {
	if node.Pipe != nil {
		result := e.Eval(node.Pipe)
		// Emit paths at the top level (ActionNode), not at every pipe level
		// This prevents duplicate emissions from nested pipes
		e.Emit(result.paths...)
		return result
	}
	return evalResult{}
}

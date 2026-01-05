package parser

import (
	"fmt"
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
	tree       *parse.Tree
	out        *path.Paths
	prefixes   path.Paths
	paramPaths map[string]*path.Path
	paramLits  map[string]string
	idx        *TemplateIndex
	inStack    map[string]bool
	depth      int
	maxDepth   int
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
		if p.ID() == "/" {
			continue
		}
		e.out.Append(p)
	}
}

// addPrefixes generates paths for all context prefixes if any are set.
// Used when creating relative paths inside with/range blocks.
// Returns a slice with the original path if no prefixes are set.
func (e *evalCtx) addPrefixes(p *path.Path) path.Paths {
	if !e.hasPrefix() {
		return path.Paths{p}
	}

	var result path.Paths
	for _, prefix := range e.prefixes {
		prefixed := prefix.Join(*p)
		result = append(result, &prefixed)
	}
	return result
}

// WithPrefixes sets context prefixes and returns a cleanup function.
// Used by with/range to change the meaning of "." in nested scopes.
// If prefixes is nil or empty, no prefix is set (no-op that still returns a valid cleanup function).
func (e *evalCtx) WithPrefixes(prefixes path.Paths) func() {
	oldPrefixes := e.prefixes
	e.prefixes = prefixes
	return func() {
		e.prefixes = oldPrefixes
	}
}

// hasPrefix returns true if any context prefixes are currently set.
func (e *evalCtx) hasPrefix() bool {
	return len(e.prefixes) > 0
}

// WithDictParams sets dict parameter mappings and returns a cleanup function.
func (e *evalCtx) WithDictParams(paths map[string]*path.Path, lits map[string]string) func() {
	oldPaths := e.paramPaths
	oldLits := e.paramLits
	e.paramPaths = paths
	e.paramLits = lits
	return func() {
		e.paramPaths = oldPaths
		e.paramLits = oldLits
	}
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
	case *parse.DotNode:
		if e.hasPrefix() {
			return evalResult{paths: e.addPrefixes(&path.Path{})}
		}
		// Outside with/range, return empty path (root context)
		return evalResult{paths: path.Paths{path.NewPath()}}
	case *parse.ChainNode:
		return e.evalChainNode(node)

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
	case *parse.VariableNode:
		return e.evalVariableNode(node)
	case *parse.IdentifierNode:
		return evalResult{}
	case *parse.NilNode:
		return evalResult{}

	default:
		slog.Warn("unsupported node type in Eval", "type", node.Type(), "nodeString", fmt.Sprintf("%T", n))
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
		if result.dict == nil && result.dictLits == nil {
			e.Emit(result.paths...)
		}
		return result
	}
	return evalResult{}
}

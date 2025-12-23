package parser

import (
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// Eval is the result of evaluating a node in analysis context.
// Path is set for .Values-based field expressions; Strings holds literal strings.
type Eval struct {
	Path    *path.Path
	Strings []string
}

// Call represents an invocation of a template function.
type Call struct {
	Name      string
	Args      []parse.Node
	Input     []string   // piped input as strings (if any)
	InputPath *path.Path // piped input as path (if available)
	Piped     bool
	Node      parse.Node // for diagnostics
}

// FnCtx carries analysis-time helpers available to template functions.
type FnCtx struct {
	a *analyzer
}

func (a *analyzer) newFnCtx() *FnCtx { return &FnCtx{a: a} }

// EvalNode evaluates a sub-node (including sub-pipelines) and returns its abstract value.
func (c *FnCtx) EvalNode(n parse.Node) Eval { return c.a.evalNode(n) }

// Emit appends a path to analysis output, applying any active prefixes.
// Currently, prefixes are not implemented; we simply append the path as-is.
func (c *FnCtx) Emit(p *path.Path) {
	if p == nil || c == nil || c.a == nil || c.a.out == nil {
		return
	}
	c.a.out.Append(p)
}

// WithPrefixes pushes one or more prefixes; returns a pop closure.
// Prefix application is a no-op for now (range/with not implemented yet).
func (c *FnCtx) WithPrefixes(_ ...*path.Path) func() { return func() {} }

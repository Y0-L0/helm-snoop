package parser

import (
	"log/slog"
	"strings"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// evalPipe evaluates a full pipeline and adds resulting .Values Paths to a.out.
func (a *analyzer) evalPipe(node *parse.PipeNode) {
	if node == nil || len(node.Cmds) == 0 {
		return
	}
	var cur interface{}
	for i, cmd := range node.Cmds {
		cur = a.evalCommandAbs(cmd, cur, i > 0)
	}
	if p := collectFromAbstract(cur); p != nil {
		a.out.Append(p)
	}
}

// evalCommandAbs evaluates one command into an abstract value.
// When piped is true, the input is appended as the last argument for functions.
func (a *analyzer) evalCommandAbs(cmd *parse.CommandNode, input interface{}, piped bool) interface{} {
	if cmd == nil || len(cmd.Args) == 0 {
		return nil
	}
	// Function call if first arg is an identifier; otherwise a literal/field.
	id, ok := cmd.Args[0].(*parse.IdentifierNode)
	if !ok {
		// Not a function: expect a single literal/field
		if len(cmd.Args) != 1 {
			slog.Warn("command with unexpected arg count", "args_len", len(cmd.Args))
			must("command with unexpected arg count")
			return nil
		}
		return a.evalArgNode(cmd.Args[0])
	}
	// Build args from remaining nodes
	args := make([]interface{}, 0, len(cmd.Args)-1+1)
	for _, aNode := range cmd.Args[1:] {
		if v := a.evalArgNode(aNode); v != nil {
			args = append(args, v)
		}
	}
	if piped && input != nil {
		args = append(args, input) // pipeline passes previous value last
	}
	if id.Ident == "include" {
		a.evalInclude(args)
		return nil
	}
	logNotImplementedCommand(a.tree, id.Ident, cmd)
	fn := getTemplateFunction(id.Ident)
	return fn(args...)
}

func (a *analyzer) evalInclude(args []interface{}) {
	if len(args) == 0 {
		slog.Warn("include: missing template name")
		must("include: missing template name")
		return
	}
	// first arg is template name, represented as KeySet from a string literal
	var name string
	switch v := args[0].(type) {
	case KeySet:
		if len(v) != 1 {
			slog.Warn("include: invalid name literal", "len", len(v))
			must("include: invalid name literal")
			return
		}
		name = v[0]
	case string:
		name = v
	default:
		slog.Warn("include: unexpected name arg type", "arg", args[0])
		must("include: unexpected name arg type")
		return
	}
	if a.idx == nil {
		slog.Warn("include: no template index set")
		must("include: no template index")
		return
	}
	def, ok := a.idx.get(name)
	if !ok {
		slog.Warn("include: unknown template name", "name", name)
		must("include: unknown template name")
		return
	}

	defer a.withIncludeScope(name)()

	oldTree := a.tree
	a.tree = def.tree
	a.collect(def.root)
	a.tree = oldTree
}

// withIncludeScope runs fn within an include expansion scope for the given template name.
// It enforces recursion and depth limits and restores state afterwards.
func (a *analyzer) withIncludeScope(name string) func() {
	if a.inStack == nil {
		a.inStack = make(map[string]bool)
	}
	if a.depth >= a.maxDepth {
		slog.Warn("include: max depth exceeded", "depth", a.depth)
		panic("include: max depth exceeded")
	}
	if a.inStack[name] {
		slog.Warn("include: recursion detected", "name", name)
		panic("include: recursion detected")
	}
	a.inStack[name] = true
	a.depth++

	return func() {
		a.depth--
		delete(a.inStack, name)
	}
}

func (a *analyzer) evalArgNode(n parse.Node) interface{} {
	switch an := n.(type) {
	case *parse.PipeNode:
		// nested pipe used as an argument; ignore for analysis for now
		return nil
	case *parse.FieldNode:
		if len(an.Ident) > 0 && an.Ident[0] == "Values" {
			if key := strings.Join(an.Ident[1:], "."); key != "" {
				return path.NewPath(an.Ident[1:]...)
			}
			// bare .Values: ignore
			return nil
		}
		// Non-.Values field like .Chart.Name, .Release.Name, etc.: ignore silently.
		return nil
	case *parse.DotNode:
		// "." as an argument (e.g., include "x" .) — ignore for analysis.
		return nil
	case *parse.VariableNode:
		// Variables (e.g., $name) are ignored in this minimal analysis phase.
		return nil
	case *parse.IdentifierNode:
		// treat non-function identifiers used as args as no-ops for analysis
		return nil
	case *parse.BoolNode:
		// literals we don't care about for analysis
		return nil
	case *parse.NumberNode:
		// numeric literal key (e.g., index with integers)
		if an.Text != "" {
			return KeySet{an.Text}
		}
		return nil
	case *parse.StringNode:
		if an.Text != "" {
			return KeySet{an.Text}
		}
	}
	slog.Warn("unsupported node kind", "node", n)
	must("evalArgNode: unsupported node kind")
	return nil
}

func collectFromAbstract(v interface{}) *path.Path {
	switch t := v.(type) {
	case *path.Path:
		return t
	case KeySet:
		// literals alone don't constitute a .Values read
		return nil
	default:
		return nil
	}
}

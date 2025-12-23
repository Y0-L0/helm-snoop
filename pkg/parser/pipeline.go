package parser

import (
	"log/slog"
	"strings"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// evalPipe evaluates a full pipeline. Functions emit paths directly.
// Returns any literal strings produced by the last command.
func (a *analyzer) evalPipe(node *parse.PipeNode) []string {
	if node == nil || len(node.Cmds) == 0 {
		return nil
	}
	var lastStrings []string
	var lastNonFunc Eval
	hadFunc := false
	for i, cmd := range node.Cmds {
		strs, nf, isFunc := a.evalCommand(cmd, lastStrings, lastNonFunc, i > 0)
		if isFunc {
			hadFunc = true
			lastStrings = strs
		} else {
			lastNonFunc = nf
			lastStrings = nf.Strings
		}
	}
	// If there was no function in the pipeline and we ended with a .Values path, emit it.
	if !hadFunc && lastNonFunc.Path != nil {
		a.out.Append(lastNonFunc.Path)
	}
	return lastStrings
}

// evalCommand evaluates one command.
// When piped is true, the input strings are appended as the last arg for functions.
// Returns (lastStrings, lastNonFuncEval, isFunc).
func (a *analyzer) evalCommand(cmd *parse.CommandNode, input []string, prevNF Eval, piped bool) ([]string, Eval, bool) {
	if cmd == nil || len(cmd.Args) == 0 {
		return nil, Eval{}, false
	}
	// Function call if first arg is an identifier; otherwise a literal/field.
	if id, ok := cmd.Args[0].(*parse.IdentifierNode); ok {
		call := Call{Name: id.Ident, Args: cmd.Args[1:], Input: nil, InputPath: nil, Piped: piped, Node: cmd}
		if piped {
			call.Input = input
			call.InputPath = prevNF.Path
		}
		fn := getTemplateFunction(id.Ident)
		ctx := a.newFnCtx()
		strs := fn(ctx, call)
		return strs, Eval{}, true
	}
	// Not a function: expect a single literal/field
	if len(cmd.Args) != 1 {
		slog.Warn("command with unexpected arg count", "args_len", len(cmd.Args), "args", cmd.Args)
		must("command with unexpected arg count")
		return nil, Eval{}, false
	}
	ev := a.evalNode(cmd.Args[0])
	return ev.Strings, ev, false
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

// evalNode evaluates an argument node and returns Eval (Path and/or literal strings).
func (a *analyzer) evalNode(n parse.Node) Eval {
	switch an := n.(type) {
	case *parse.PipeNode:
		return Eval{Strings: a.evalPipe(an)}
	case *parse.CommandNode:
		// Support nested path synthesis for index/get used as sub-commands.
		if len(an.Args) > 0 {
			if id, ok := an.Args[0].(*parse.IdentifierNode); ok {
				switch id.Ident {
				case "index":
					// Expect at least base; keys optional (can be piped later).
					if len(an.Args) < 2 {
						return Eval{}
					}
					base := a.evalNode(an.Args[1]).Path
					if base == nil {
						return Eval{}
					}
					p := *base
					for _, arg := range an.Args[2:] {
						lits := a.evalNode(arg).Strings
						if len(lits) != 1 {
							return Eval{}
						}
						p = p.WithAny(lits[0])
					}
					return Eval{Path: &p}
				case "get":
					if len(an.Args) != 3 { // get base key
						return Eval{}
					}
					base := a.evalNode(an.Args[1]).Path
					if base == nil {
						return Eval{}
					}
					lits := a.evalNode(an.Args[2]).Strings
					if len(lits) != 1 {
						return Eval{}
					}
					p := base.WithAny(lits[0])
					return Eval{Path: &p}
				}
			}
		}
		return Eval{}
	case *parse.FieldNode:
		if len(an.Ident) > 0 && an.Ident[0] == "Values" {
			if key := strings.Join(an.Ident[1:], "."); key != "" {
				return Eval{Path: path.NewPath(an.Ident[1:]...)}
			}
			return Eval{}
		}
		return Eval{}
	case *parse.DotNode:
		return Eval{}
	case *parse.VariableNode:
		return Eval{}
	case *parse.IdentifierNode:
		return Eval{}
	case *parse.BoolNode:
		return Eval{}
	case *parse.NumberNode:
		if an.Text != "" {
			return Eval{Strings: []string{an.Text}}
		}
		return Eval{}
	case *parse.StringNode:
		if an.Text != "" {
			return Eval{Strings: []string{an.Text}}
		}
		return Eval{}
	default:
		slog.Warn("unsupported node kind in evalNode", "node", n)
		must("evalNode: unsupported node kind")
		return Eval{}
	}
}

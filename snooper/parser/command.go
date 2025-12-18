package parser

import (
	"strings"
	"text/template/parse"
)

// evaluatePipe evaluates a full pipeline and returns resulting .Values paths.
func evaluatePipe(node *parse.PipeNode) []string {
	if node == nil || len(node.Cmds) == 0 {
		return nil
	}
	var cur interface{}
	for i, cmd := range node.Cmds {
		cur = evalCommandAbs(cmd, cur, i > 0)
	}
	return collectFromAbstract(cur)
}

// evalCommandAbs evaluates one command into an abstract value.
// When piped is true, the input is appended as the last argument for functions.
func evalCommandAbs(cmd *parse.CommandNode, input interface{}, piped bool) interface{} {
	if cmd == nil || len(cmd.Args) == 0 {
		return nil
	}
	// Function call if first arg is an identifier; otherwise a literal/field.
	id, ok := cmd.Args[0].(*parse.IdentifierNode)
	if !ok {
		// Not a function: expect a single literal/field
		if len(cmd.Args) != 1 {
			panic("not implemented")
		}
		return evalArgNode(cmd.Args[0])
	}
	// Build args from remaining nodes
	args := make([]interface{}, 0, len(cmd.Args)-1+1)
	for _, a := range cmd.Args[1:] {
		if v := evalArgNode(a); v != nil {
			args = append(args, v)
		}
	}
	if piped && input != nil {
		args = append(args, input) // pipeline passes previous value last
	}
	fn := getTemplateFunction(id.Ident)
	return fn(args...)
}

func evalArgNode(n parse.Node) interface{} {
	switch a := n.(type) {
	case *parse.FieldNode:
		if len(a.Ident) > 0 && a.Ident[0] == "Values" {
			if key := strings.Join(a.Ident[1:], "."); key != "" {
				return AbsPath{Segs: a.Ident[1:]}
			}
			// bare .Values: ignore
			return nil
		}
	case *parse.IdentifierNode:
		// treat non-function identifiers used as args as no-ops for analysis
		return nil
	case *parse.BoolNode:
		// literals we don't care about for analysis
		return nil
	case *parse.NumberNode:
		// numeric literal key (e.g., index with integers)
		if a.Text != "" {
			return LiteralSet{Values: []string{a.Text}}
		}
		return nil
	case *parse.StringNode:
		if a.Text != "" {
			return LiteralSet{Values: []string{a.Text}}
		}
	}
	panic("not implemented")
}

func collectFromAbstract(v interface{}) []string {
	switch t := v.(type) {
	case AbsPath:
		return []string{strings.Join(t.Segs, ".")}
	case LiteralSet:
		// literals alone don't constitute a .Values read
		return nil
	default:
		return nil
	}
}

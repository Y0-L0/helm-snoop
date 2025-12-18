package parser

import (
	"strings"
	"text/template/parse"
)

// evaluateCommand extracts .Values reads from a single command.
// It implements small per-function behaviors and keeps everything static.
func evaluateCommand(cmd *parse.CommandNode) []string {
	if cmd == nil || len(cmd.Args) == 0 {
		return nil
	}
	// Evaluate arguments into abstract values
	args := make([]interface{}, 0, len(cmd.Args))
	name := ""
	for i, a := range cmd.Args {
		if i == 0 {
			if id, ok := a.(*parse.IdentifierNode); ok {
				name = id.Ident
				continue // do not treat identifier as argument
			}
		}
		args = append(args, evalArgNode(a))
	}
	if name == "" {
		// No function: collect fields directly from args
		return scanFieldArgs(cmd.Args)
	}
	// Lookup function in funcMap; execute to get abstract result
	fm := newFuncMap()
	if fn, ok := fm[name]; ok {
		if f, ok := fn.(func(...interface{}) interface{}); ok {
			res := f(args...)
			return collectFromAbstract(res)
		}
	}
	// Unknown function: best-effort direct scan
	return scanFieldArgs(cmd.Args[1:])
}

func evalArgNode(n parse.Node) interface{} {
	switch a := n.(type) {
	case *parse.FieldNode:
		if len(a.Ident) > 0 && a.Ident[0] == "Values" {
			if key := strings.Join(a.Ident[1:], "."); key != "" {
				return AbsPath{Segs: a.Ident[1:]}
			}
		}
	case *parse.StringNode:
		if a.Text != "" {
			return LiteralSet{Values: []string{a.Text}}
		}
	}
	return Unknown{}
}

func collectFromAbstract(v interface{}) []string {
	switch t := v.(type) {
	case AbsPath:
		return []string{strings.Join(t.Segs, ".")}
	case LiteralSet:
		// literals alone don't constitute a .Values read
		return nil
	case Unknown:
		return nil
	default:
		return nil
	}
}

func scanFieldArgs(args []parse.Node) []string {
	if len(args) == 0 {
		return nil
	}
	var out []string
	for _, a := range args {
		out = append(out, collectFieldFromNode(a)...)
	}
	return out
}

func collectFieldFromNode(n parse.Node) []string {
	switch a := n.(type) {
	case *parse.FieldNode:
		if len(a.Ident) > 0 && a.Ident[0] == "Values" {
			if key := strings.Join(a.Ident[1:], "."); key != "" {
				return []string{key}
			}
		}
	}
	return nil
}

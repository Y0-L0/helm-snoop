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
	// First arg must be an identifier to be treated as a function name.
	id, ok := cmd.Args[0].(*parse.IdentifierNode)
	if !ok {
		return scanFieldArgs(cmd.Args)
	}

	args := make([]interface{}, 0, len(cmd.Args)-1)
	for _, a := range cmd.Args[1:] {
		args = append(args, evalArgNode(a))
	}

	function := getTemplateFunction(id.Ident)
	result := function(args...)
	return collectFromAbstract(result)
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

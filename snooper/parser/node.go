package parser

import (
	"text/template/parse"
)

// collectUsedValues walks a template node and returns all direct .Values paths
// observed as dot-joined strings (e.g., "config.message").
func collectUsedValues(node parse.Node) []string {
	if node == nil {
		return nil
	}
	out := make([]string, 0)
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return out
		}
		for _, nd := range n.Nodes {
			out = append(out, collectUsedValues(nd)...)
		}
	case *parse.ActionNode:
		if n.Pipe == nil {
			return out
		}
		out = append(out, evaluatePipe(n.Pipe)...)
	case *parse.IfNode:
		panic("`if` / `else` is not implemented")
		// out = append(out, collectUsedValues(n.List)...)
		// out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.WithNode:
		panic("`with` is not implemented")
		// out = append(out, collectUsedValues(n.List)...)
		// out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.RangeNode:
		panic("`range` is not implemented")
		// out = append(out, collectUsedValues(n.List)...)
		// out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.TemplateNode:
		panic("`template` is not implemented")
		// include/template resolution not implemented yet
	default:
		_ = n
	}
	return out
}

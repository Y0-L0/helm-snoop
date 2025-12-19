package parser

import (
	"log/slog"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/snooper/path"
)

// collectUsedValues walks a template node and returns all direct .Values paths
// observed as dot-joined strings (e.g., "config.message").
func collectUsedValues(node parse.Node, out *path.Paths) {
	if node == nil {
		return
	}
	slog.Debug("Handling node", "node", node)
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return
		}
		slog.Debug("Expanding list nodes", "nodes", n.Nodes)
		for _, nd := range n.Nodes {
			collectUsedValues(nd, out)
		}
	case *parse.ActionNode:
		if n.Pipe == nil {
			return
		}
		evaluatePipe(n.Pipe, out)
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
}

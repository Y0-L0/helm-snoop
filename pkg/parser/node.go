package parser

import (
	"log/slog"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// collectUsedValues walks a template node and returns all direct .Values Paths
func collectUsedValues(rawNode parse.Node, out *path.Paths) {
	if rawNode == nil {
		return
	}
	slog.Debug("Handling node", "node", rawNode)
	switch node := rawNode.(type) {
	case *parse.ListNode:
		if node == nil {
			return
		}
		slog.Debug("Expanding list nodes", "nodes", node.Nodes)
		for _, childNode := range node.Nodes {
			collectUsedValues(childNode, out)
		}
	case *parse.ActionNode:
		if node.Pipe == nil {
			return
		}
		evalPipe(node.Pipe, out)
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
		_ = node
	}
}

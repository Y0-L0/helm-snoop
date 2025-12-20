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
		slog.Warn("if/else not implemented", "node", node)
		must("if/else not implemented")
		return
		// out = append(out, collectUsedValues(n.List)...)
		// out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.WithNode:
		slog.Warn("with not implemented", "node", node)
		must("with not implemented")
		return
		// out = append(out, collectUsedValues(n.List)...)
		// out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.RangeNode:
		slog.Warn("range not implemented", "node", node)
		must("range not implemented")
		return
		// out = append(out, collectUsedValues(n.List)...)
		// out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.TemplateNode:
		slog.Warn("template action not implemented", "node", node)
		must("template action not implemented")
		return
		// include/template resolution not implemented yet
	default:
		_ = node
	}
}

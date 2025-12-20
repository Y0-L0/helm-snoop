package parser

import (
	"log/slog"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// collectUsedValues walks a template node and returns all direct .Values Paths
func collectUsedValues(tree *parse.Tree, rawNode parse.Node, out *path.Paths) {
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
			collectUsedValues(tree, childNode, out)
		}
	case *parse.ActionNode:
		if node.Pipe == nil {
			return
		}
		evalPipe(tree, node.Pipe, out)
	case *parse.IfNode:
		loc, _ := tree.ErrorContext(node)
		slog.Warn("if/else not implemented", "pipe", node.Pipe, "pos", loc)
		must("if/else not implemented")
		return
		// out = append(out, collectUsedValues(n.List)...)
		// out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.WithNode:
		loc, _ := tree.ErrorContext(node)
		slog.Warn("with not implemented", "pipe", node.Pipe, "pos", loc)
		must("with not implemented")
		return
		// out = append(out, collectUsedValues(n.List)...)
		// out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.RangeNode:
		loc, _ := tree.ErrorContext(node)
		slog.Warn("range not implemented", "pipe", node.Pipe, "pos", loc)
		must("range not implemented")
		return
		// out = append(out, collectUsedValues(n.List)...)
		// out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.TemplateNode:
		loc, _ := tree.ErrorContext(node)
		slog.Warn("template action not implemented", "name", node.Name, "pipe", node.Pipe, "pos", loc)
		must("template action not implemented")
		return
		// include/template resolution not implemented yet
	default:
		_ = node
	}
}

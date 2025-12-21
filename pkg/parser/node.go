package parser

import (
	"log/slog"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// analyzer holds traversal state for parsing a single template tree.
// It is intentionally unexported and constructed by parseFile/GetUsages.
type analyzer struct {
	tree *parse.Tree
	out  *path.Paths
}

// collect walks a template node and appends observed .Values paths to a.out.
func (a *analyzer) collect(rawNode parse.Node) {
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
			a.collect(childNode)
		}
	case *parse.ActionNode:
		if node.Pipe == nil {
			return
		}
		a.evalPipe(node.Pipe)
	case *parse.IfNode:
		// record values used in the condition
		a.evalPipe(node.Pipe)
		// evaluate both branches
		if node.List != nil {
			a.collect(node.List)
		}
		if node.ElseList != nil {
			a.collect(node.ElseList)
		}
		return
	case *parse.WithNode:
		loc, _ := a.tree.ErrorContext(node)
		slog.Warn("with not implemented", "pipe", node.Pipe, "pos", loc)
		must("with not implemented")
		return
		// range/with body handling to be implemented later
	case *parse.RangeNode:
		loc, _ := a.tree.ErrorContext(node)
		slog.Warn("range not implemented", "pipe", node.Pipe, "pos", loc)
		must("range not implemented")
		return
	case *parse.TemplateNode:
		loc, _ := a.tree.ErrorContext(node)
		slog.Warn("template action not implemented", "name", node.Name, "pipe", node.Pipe, "pos", loc)
		must("template action not implemented")
		return
	default:
		_ = node
	}
}

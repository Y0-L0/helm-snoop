package analyzer

import (
	"fmt"
	"strings"
	"text/template/parse"

	chart "helm.sh/helm/v4/pkg/chart/v2"
)

// getUsages walks all chart templates and returns a flat list of observed .Values paths.
func getUsages(ch *chart.Chart) ([]string, error) {
	all := make([]string, 0)
	for _, tmpl := range ch.Templates {
		vals, err := parseFile(tmpl.Name, tmpl.Data)
		if err != nil {
			return nil, err
		}
		all = append(all, vals...)
	}
	return all, nil
}

// parseFile parses one template file and returns all observed .Values paths.
func parseFile(name string, data []byte) ([]string, error) {
	trees, err := parse.Parse(name, string(data), "", "", nil)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}
	out := make([]string, 0)
	for _, t := range trees {
		out = append(out, collectUsedValues(t.Root)...)
	}
	return out, nil
}

// collectUsedValues walks a template node and returns all direct .Values paths
// observed as dot-joined strings (e.g., "config.message").
func collectUsedValues(node parse.Node) []string {
	if node == nil {
		return nil
	}
	out := make([]string, 0)
	switch n := node.(type) {
	case *parse.ListNode:
		if n != nil {
			for _, nd := range n.Nodes {
				out = append(out, collectUsedValues(nd)...)
			}
		}
	case *parse.ActionNode:
		if n.Pipe != nil {
			for _, cmd := range n.Pipe.Cmds {
				for _, arg := range cmd.Args {
					switch a := arg.(type) {
					case *parse.FieldNode:
						if len(a.Ident) > 0 && a.Ident[0] == "Values" {
							if key := strings.Join(a.Ident[1:], "."); key != "" {
								out = append(out, key)
							}
						}
					}
				}
			}
		}
	case *parse.IfNode:
		out = append(out, collectUsedValues(n.List)...)
		out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.WithNode:
		out = append(out, collectUsedValues(n.List)...)
		out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.RangeNode:
		out = append(out, collectUsedValues(n.List)...)
		out = append(out, collectUsedValues(n.ElseList)...)
	case *parse.TemplateNode:
		// include/template resolution not implemented yet
	default:
		_ = n
	}
	return out
}

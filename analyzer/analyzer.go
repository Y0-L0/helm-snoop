package analyzer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template/parse"

	"gopkg.in/yaml.v3"
)

// Analyze performs a minimal analysis (MVP Stage 1):
// - parses templates under <chartPath>/templates (non-recursive for MVP)
// - extracts direct .Values.foo.bar field chains
// - flattens values.yaml to defined keys
// - returns referenced, defined-not-used, and used-not-defined (all sorted, unique)
func Analyze(chartPath string, res *Result) error {
	tmplDir := filepath.Join(chartPath, "templates")

	usedSet := map[string]struct{}{}

	// Read templates (MVP: top-level files only)
	entries, readErr := os.ReadDir(tmplDir)
	if readErr != nil {
		return fmt.Errorf("read templates dir: %w", readErr)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "NOTES.txt" {
			continue
		}
		// Limit to simple YAML-like files for MVP
		if !hasYAMLExt(name) && !strings.HasSuffix(name, ".tpl") {
			continue
		}
		path := filepath.Join(tmplDir, name)
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return fmt.Errorf("read template %s: %w", path, rerr)
		}
		trees, perr := parse.Parse(path, string(data), "", "", nil)
		if perr != nil {
			return fmt.Errorf("parse template %s: %w", path, perr)
		}
		for _, t := range trees {
			collectUsedValues(t.Root, usedSet)
		}
	}

	// Flatten values.yaml
	definedSet := map[string]struct{}{}
	vyPath := filepath.Join(chartPath, "values.yaml")
	if _, statErr := os.Stat(vyPath); statErr == nil {
		vdata, ferr := os.ReadFile(vyPath)
		if ferr != nil {
			return fmt.Errorf("read values.yaml: %w", ferr)
		}
		var v interface{}
		if yerr := yaml.Unmarshal(vdata, &v); yerr != nil {
			return fmt.Errorf("parse values.yaml: %w", yerr)
		}
		flattenValues("", v, definedSet)
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("stat values.yaml: %w", statErr)
	}

	// Build slices
	res.Referenced = setToSortedSlice(usedSet)

	// Compute diffs
	res.DefinedNotUsed = diffSets(definedSet, usedSet)
	res.UsedNotDefined = diffSets(usedSet, definedSet)

	sort.Strings(res.DefinedNotUsed)
	sort.Strings(res.UsedNotDefined)

	return nil
}

func hasYAMLExt(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml")
}

func collectUsedValues(node parse.Node, out map[string]struct{}) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return
		}
		for _, nd := range n.Nodes {
			collectUsedValues(nd, out)
		}
	case *parse.ActionNode:
		if n.Pipe != nil {
			for _, cmd := range n.Pipe.Cmds {
				for _, arg := range cmd.Args {
					switch a := arg.(type) {
					case *parse.FieldNode:
						// .Values.foo.bar => Ident: ["Values","foo","bar"]
						if len(a.Ident) > 0 && a.Ident[0] == "Values" {
							if key := strings.Join(a.Ident[1:], "."); key != "" {
								out[key] = struct{}{}
							}
						}
					}
				}
			}
		}
	case *parse.IfNode:
		collectUsedValues(n.List, out)
		collectUsedValues(n.ElseList, out)
	case *parse.WithNode:
		collectUsedValues(n.List, out)
		collectUsedValues(n.ElseList, out)
	case *parse.RangeNode:
		collectUsedValues(n.List, out)
		collectUsedValues(n.ElseList, out)
	case *parse.TemplateNode:
		// MVP: do not resolve; body may be nil here
	default:
		// ignore: TextNode, etc.
		_ = n
	}
}

func flattenValues(prefix string, v interface{}, out map[string]struct{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, child := range val {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenValues(key, child, out)
		}
	case map[interface{}]interface{}:
		for rk, child := range val {
			ks, ok := rk.(string)
			if !ok {
				continue
			}
			key := ks
			if prefix != "" {
				key = prefix + "." + ks
			}
			flattenValues(key, child, out)
		}
	case []interface{}:
		// arrays: mark container key only (leaf value indices are not keys)
		if prefix != "" {
			out[prefix] = struct{}{}
		}
		// do not descend into elements for key generation
	default:
		if prefix != "" {
			out[prefix] = struct{}{}
		}
	}
}

func setToSortedSlice(m map[string]struct{}) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func diffSets(a, b map[string]struct{}) []string {
	// return keys in a that are not in b
	if len(a) == 0 {
		return nil
	}
	out := make([]string, 0)
	for k := range a {
		if _, ok := b[k]; !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// helper for tests to ensure fs semantics are consistent (unused currently)
var _ fs.FileInfo

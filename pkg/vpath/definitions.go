package vpath

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/y0-l0/helm-snoop/internal/assert"
)

// GetDefinitions unmarshals raw YAML and returns all leaf definition paths
// with file location context (line/column) attached.
func GetDefinitions(data []byte, fileName string) (Paths, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	d := defCtx{fileName: fileName}
	d.eval(Path{}, &node)
	return d.out, nil
}

type defCtx struct {
	fileName string
	out      Paths
}

func (d *defCtx) emit(path Path) {
	if len(path.Contexts) == 0 {
		assert.Must("defCtx.emit: path has no context")
	}
	p := path
	d.out = append(d.out, &p)
}

func (d *defCtx) eval(path Path, node *yaml.Node) {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			d.eval(path, child)
		}
	case yaml.MappingNode:
		if len(node.Content) == 0 {
			d.emit(path)
			return
		}
		for i := 0; i < len(node.Content)-1; i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			childPath := path.WithKey(keyNode.Value)
			childPath.Contexts = Contexts{{FileName: d.fileName, Line: keyNode.Line, Column: keyNode.Column}}
			d.eval(childPath, valNode)
		}
	case yaml.SequenceNode:
		if len(node.Content) == 0 {
			d.emit(path)
			return
		}
		for i, child := range node.Content {
			d.eval(path.WithIdx(strconv.Itoa(i)), child)
		}
	case yaml.ScalarNode:
		d.emit(path)
	case yaml.AliasNode:
		d.eval(path, node.Alias)
	}
}

// GetDefinitionsFromMap walks an already-parsed map and returns all leaf
// definition paths. No line/column context is available; paths get a
// source-only context using the provided source label.
func GetDefinitionsFromMap(m map[string]any, source string) Paths {
	if len(m) == 0 {
		return nil
	}
	d := defCtx{fileName: source}
	d.evalMap(Path{}, m)
	return d.out
}

func (d *defCtx) evalMap(path Path, m map[string]any) {
	if len(m) == 0 {
		d.emitSourceOnly(path)
		return
	}
	for key, val := range m {
		child := path.WithKey(key)
		child.Contexts = Contexts{{FileName: d.fileName}}
		d.evalAny(child, val)
	}
}

func (d *defCtx) evalAny(path Path, val any) {
	switch v := val.(type) {
	case map[string]any:
		d.evalMap(path, v)
	case []any:
		if len(v) == 0 {
			d.emitSourceOnly(path)
			return
		}
		for i, item := range v {
			d.evalAny(path.WithIdx(strconv.Itoa(i)), item)
		}
	default:
		d.emitSourceOnly(path)
	}
}

func (d *defCtx) emitSourceOnly(path Path) {
	if len(path.Contexts) == 0 {
		path.Contexts = Contexts{{FileName: d.fileName}}
	}
	p := path
	d.out = append(d.out, &p)
}

// FormatMapSource formats a source label for inline extraValues from a config file.
func FormatMapSource(configPath, chartKey string) string {
	if chartKey == "" {
		return fmt.Sprintf("%s:global.extraValues", configPath)
	}
	return fmt.Sprintf("%s:charts.%s.extraValues", configPath, chartKey)
}

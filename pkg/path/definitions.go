package path

import (
	"strconv"

	"github.com/y0-l0/helm-snoop/internal/assert"
	"gopkg.in/yaml.v3"
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
			childPath.Contexts = []PathContext{{FileName: d.fileName, Line: keyNode.Line, Column: keyNode.Column}}
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

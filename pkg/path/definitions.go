package path

import (
	"fmt"
	"log/slog"
	"strconv"
)

func GetDefinitions(path Path, abstractNode interface{}, out *Paths) {
	slog.Debug("GetDefinitions called", "path", path, "v", abstractNode)
	switch node := abstractNode.(type) {
	// object/map with string keys
	case map[string]interface{}:
		if len(node) == 0 {
			*out = append(*out, &path)
			return
		}
		for key, child := range node {
			GetDefinitions(path.WithKey(key), child, out)
		}
	// object/map with non-string keys
	case map[interface{}]interface{}:
		if len(node) == 0 {
			*out = append(*out, &path)
			return
		}
		for rawKey, child := range node {
			key := fmt.Sprintf("%v", rawKey)
			// treat non-string map keys as regular keys represented as strings
			GetDefinitions(path.WithKey(key), child, out)
		}
	// array/list
	case []interface{}:
		if len(node) == 0 {
			*out = append(*out, &path)
			return
		}
		// arrays: include each element index as a path segment and descend
		for i, child := range node {
			GetDefinitions(path.WithIdx(strconv.Itoa(i)), child, out)
		}
	// scalar leaf node
	default:
		slog.Debug("Found scalar leaf node. Append path to out", "node", node, "path", path)
		*out = append(*out, &path)
	}
}

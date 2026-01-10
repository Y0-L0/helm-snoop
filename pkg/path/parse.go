package path

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePath parses a path string into a Path.
// Path syntax:
//   - Must start with /
//   - Segments separated by /
//   - Integer segments (e.g., "0", "123") become anyKind
//   - "*" segments become wildcardKind
//   - Other segments become keyKind
//
// Examples:
//
//	"/image/tag"      → Key("image").Key("tag")
//	"/items/0"        → Key("items").Any("0")
//	"/config/*"       → Key("config").Wildcard()
//	"/a/*/c"          → Key("a").Wildcard().Key("c")
func ParsePath(pattern string) (*Path, error) {
	// Validate empty
	if pattern == "" {
		return nil, fmt.Errorf("empty pattern")
	}

	// Validate leading slash
	if !strings.HasPrefix(pattern, "/") {
		return nil, fmt.Errorf("pattern must start with /")
	}

	// Only slash
	if pattern == "/" {
		return nil, fmt.Errorf("empty pattern after /")
	}

	// Remove leading slash and split
	trimmed := strings.TrimPrefix(pattern, "/")

	// Check trailing slash
	if strings.HasSuffix(trimmed, "/") {
		return nil, fmt.Errorf("pattern must not have trailing slash")
	}

	segments := strings.Split(trimmed, "/")

	// Build path
	p := &Path{
		tokens: make([]string, 0, len(segments)),
		kinds:  make([]kind, 0, len(segments)),
	}

	for i, segment := range segments {
		// Check empty segment (double slash)
		if segment == "" {
			return nil, fmt.Errorf("empty segment at position %d", i)
		}

		// Determine kind
		if segment == "*" {
			p.tokens = append(p.tokens, segment)
			p.kinds = append(p.kinds, wildcardKind)
		} else if isInteger(segment) {
			// Integer → anyKind
			p.tokens = append(p.tokens, segment)
			p.kinds = append(p.kinds, anyKind)
		} else {
			// Regular key
			p.tokens = append(p.tokens, segment)
			p.kinds = append(p.kinds, keyKind)
		}
	}

	return p, nil
}

// isInteger returns true if s represents a non-negative integer
func isInteger(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

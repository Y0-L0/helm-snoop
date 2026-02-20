package path

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePath parses a path string into a Path.
// Path syntax:
//   - Optional leading dot (matches ID() output format for easy copy-paste)
//   - Segments separated by .
//   - Dots inside a segment name can be escaped as ~.
//   - Integer segments (e.g., "0", "123") become anyKind
//   - "*" segments become wildcardKind
//   - Other segments become keyKind
//
// Examples:
//
//	".image.tag"       → Key("image").Key("tag")
//	"image.tag"        → Key("image").Key("tag")
//	".items.0"         → Key("items").Any("0")
//	".config.*"        → Key("config").Wildcard()
//	".a.*.c"           → Key("a").Wildcard().Key("c")
func ParsePath(pattern string) (*Path, error) {
	// Validate empty
	if pattern == "" {
		return nil, fmt.Errorf("empty pattern")
	}

	// Reject old slash notation for a clear error message
	if strings.HasPrefix(pattern, "/") {
		return nil, fmt.Errorf("use dot notation (e.g. .image.tag), not slash notation (%s)", pattern)
	}

	// Strip optional leading dot to match ID() output format
	if strings.HasPrefix(pattern, ".") {
		pattern = pattern[1:]
	}

	// Nothing left after stripping dot
	if pattern == "" {
		return nil, fmt.Errorf("empty pattern")
	}

	segments := splitOnDots(pattern)

	// Build path
	p := &Path{
		tokens: make([]string, 0, len(segments)),
		kinds:  make([]kind, 0, len(segments)),
	}

	for i, segment := range segments {
		// Check empty segment (double dot or trailing dot)
		if segment == "" {
			if i == len(segments)-1 {
				return nil, fmt.Errorf("pattern must not have trailing dot")
			}
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

// splitOnDots splits s on unescaped dots.
// A dot preceded by ~ is an escaped dot (~.) and is not treated as a separator.
// ~~ is an escaped tilde, so ~~. splits at the dot.
func splitOnDots(s string) []string {
	var segments []string
	var current strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '~' && i+1 < len(s) {
			// Consume escape sequence (~~ or ~.)
			current.WriteByte(s[i])
			current.WriteByte(s[i+1])
			i += 2
		} else if s[i] == '.' {
			segments = append(segments, current.String())
			current.Reset()
			i++
		} else {
			current.WriteByte(s[i])
			i++
		}
	}
	segments = append(segments, current.String())
	return segments
}

// isInteger returns true if s represents a non-negative integer
func isInteger(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

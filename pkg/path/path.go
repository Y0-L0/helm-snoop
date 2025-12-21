package path

import (
	"slices"
	"strings"
)

type kind byte

const (
	indexKind kind = 'I'
	keyKind   kind = 'K'
	anyKind   kind = 'A'
)

type Path struct {
	tokens []string
	kinds  []kind
}

// Stable JsonPointer representation of the Path.
// Used for comparison and map keys.
// Does not distinguish between map (key) and array (int) segments.
// That requires comparing the kinds.
func (p Path) ID() string {
	return "/" + strings.Join(p.tokens, "/")
}

func (p Path) Compare(other Path) int {
	if c := slices.Compare(p.tokens, other.tokens); c != 0 {
		return c
	}
	return slices.Compare(p.kinds, other.kinds)
}

// KindsString returns a slash-separated string encoding of the path segment kinds.
// Example: "K/K/I/K" for /a/b/1/c
func (p Path) KindsString() string {
	if len(p.kinds) == 0 {
		return ""
	}
	// Prepend a leading slash to mirror ID().
	b := make([]byte, 0, len(p.kinds)*2)
	for _, k := range p.kinds {
		b = append(b, '/')
		switch k {
		case keyKind:
			b = append(b, 'K')
		case indexKind:
			b = append(b, 'I')
		case anyKind:
			b = append(b, 'A')
		default:
			panic("invalid kind: zero or unknown value")
		}
	}
	return string(b)
}

var escaper = strings.NewReplacer("~", "~0", "/", "~1")

func (p Path) WithKey(key string) Path {
	p.tokens = append([]string(nil), p.tokens...)
	p.tokens = append(p.tokens, escaper.Replace(key))

	p.kinds = append([]kind(nil), p.kinds...)
	p.kinds = append(p.kinds, keyKind)

	return p
}

// Key is a mutator: it appends a map key segment to the receiver Path in place.
// Prefer the immutable-style WithKey in traversal code to avoid slice aliasing across siblings.
func (p *Path) Key(key string) *Path {
	p.tokens = append(p.tokens, escaper.Replace(key))
	p.kinds = append(p.kinds, keyKind)
	return p
}

func (p Path) WithIdx(key string) Path {
	p.tokens = append([]string(nil), p.tokens...)
	p.tokens = append(p.tokens, escaper.Replace(key))

	p.kinds = append([]kind(nil), p.kinds...)
	p.kinds = append(p.kinds, indexKind)

	return p
}

// Idx is a mutator: it appends an index segment to the receiver Path in place.
// Prefer the immutable-style WithIdx in traversal code to avoid slice aliasing across siblings.
func (p *Path) Idx(key string) *Path {
	p.tokens = append(p.tokens, escaper.Replace(key))
	p.kinds = append(p.kinds, indexKind)
	return p
}

func (p Path) WithAny(token string) Path {
	p.tokens = append([]string(nil), p.tokens...)
	p.tokens = append(p.tokens, escaper.Replace(token))

	p.kinds = append([]kind(nil), p.kinds...)
	p.kinds = append(p.kinds, anyKind)

	return p
}

// Any is a mutator: it appends an unknown-kind segment to the receiver Path in place.
// Prefer the immutable-style WithAny in traversal code to avoid slice aliasing across siblings.
func (p *Path) Any(token string) *Path {
	p.tokens = append(p.tokens, escaper.Replace(token))
	p.kinds = append(p.kinds, anyKind)
	return p
}

func NewPath(tokens ...string) *Path {
	p := Path{
		make([]string, 0, len(tokens)),
		make([]kind, 0, len(tokens)),
	}
	for _, t := range tokens {
		p.Key(t)
	}
	return &p
}

// newPath helper for test Path construction
func np() *Path {
	return &Path{}
}

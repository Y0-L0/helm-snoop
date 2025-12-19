package parser

// Minimal abstract domain for function evaluation.

// LiteralSet is a finite set of string literals (used for keys).
type LiteralSet struct{ Values []string }

// Unknown marks values we cannot resolve statically.
type Unknown struct{}

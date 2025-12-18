package parser

// Minimal abstract domain for function evaluation.

// AbsPath represents a concrete .Values path as segments.
type AbsPath struct{ Segs []string }

// LiteralSet is a finite set of string literals (used for keys).
type LiteralSet struct{ Values []string }

// Unknown marks values we cannot resolve statically.
type Unknown struct{}

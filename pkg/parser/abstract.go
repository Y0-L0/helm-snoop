package parser

// Minimal abstract domain for function evaluation.

// KeySet is a finite set of literal keys (used as operands to index/get).
type KeySet []string

// Unknown marks values we cannot resolve statically.
type Unknown struct{}

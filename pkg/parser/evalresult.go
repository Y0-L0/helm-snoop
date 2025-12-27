package parser

import (
	"github.com/y0-l0/helm-snoop/pkg/path"
)

type evalResult struct {
	// args contains literal string values extracted from the node.
	// Used for literal folding and as keys for path synthesis (index, get).
	args []string

	// paths contains the union of all .Values paths discovered during evaluation.
	paths []*path.Path

	// dict provides structure tracking for dict literals.
	// Maps literal keys to their corresponding .Values paths.
	// Only populated by the dict function; nil otherwise.
	// Functions like index, get, and include can use this for precise resolution.
	dict map[string]*path.Path
}

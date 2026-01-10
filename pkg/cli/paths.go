package cli

import (
	"fmt"
	"strings"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// cliPaths is a path.Paths that implements pflag.Value for CLI parsing.
type cliPaths path.Paths

// String returns a string representation of the paths (for pflag)
func (p *cliPaths) String() string {
	if p == nil || len(*p) == 0 {
		return ""
	}
	ids := make([]string, len(*p))
	for i, path := range *p {
		ids[i] = path.ID()
	}
	return "[" + strings.Join(ids, ", ") + "]"
}

// Set parses and validates a single path (called by pflag)
func (p *cliPaths) Set(value string) error {
	parsedPath, err := path.ParsePath(value)
	if err != nil {
		return fmt.Errorf("invalid ignore path %q: %w", value, err)
	}

	*p = append(*p, parsedPath)
	return nil
}

// Type returns the type name for help text
func (p *cliPaths) Type() string {
	return "ignorePath"
}

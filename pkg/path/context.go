package path

import (
	"fmt"
	"slices"
)

type PathContext struct {
	FileName     string
	TemplateName string
	Line         int
	Column       int
}

// Contexts is a slice of PathContext with helper methods.
type Contexts []PathContext

// Deduplicate removes duplicate entries, preserving order.
func (cs Contexts) Deduplicate() Contexts {
	if len(cs) == 0 {
		return nil
	}
	out := make(Contexts, 0, len(cs))
	for _, c := range cs {
		if !slices.Contains(out, c) {
			out = append(out, c)
		}
	}
	return out
}

func (c PathContext) String() string {
	if c.TemplateName != "" {
		return fmt.Sprintf("%s:%d:%d (%s)", c.FileName, c.Line, c.Column, c.TemplateName)
	}
	return fmt.Sprintf("%s:%d:%d", c.FileName, c.Line, c.Column)
}

func (c PathContext) ToJSON() PathContextJSON {
	return PathContextJSON{
		File:     c.FileName,
		Template: c.TemplateName,
		Line:     c.Line,
		Column:   c.Column,
	}
}

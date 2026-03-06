package vpath

import (
	"fmt"
	"slices"
)

type Context struct {
	FileName     string
	TemplateName string
	Line         int
	Column       int
}

// Contexts is a slice of Context with helper methods.
type Contexts []Context

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

func (c Context) String() string {
	if c.TemplateName != "" {
		return fmt.Sprintf("%s:%d:%d (%s)", c.FileName, c.Line, c.Column, c.TemplateName)
	}
	return fmt.Sprintf("%s:%d:%d", c.FileName, c.Line, c.Column)
}

func (c Context) ToJSON() ContextJSON {
	return ContextJSON{
		File:     c.FileName,
		Template: c.TemplateName,
		Line:     c.Line,
		Column:   c.Column,
	}
}

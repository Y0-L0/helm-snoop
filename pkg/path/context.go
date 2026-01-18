package path

import "fmt"

type PathContext struct {
	FileName     string
	TemplateName string
	Line         int
	Column       int
}

func (c PathContext) String() string {
	if c.TemplateName != "" {
		return fmt.Sprintf("%s:%d:%d (%s)", c.FileName, c.Line, c.Column, c.TemplateName)
	}
	return fmt.Sprintf("%s:%d:%d", c.FileName, c.Line, c.Column)
}

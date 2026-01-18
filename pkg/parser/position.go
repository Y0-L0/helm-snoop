package parser

import "strings"

// CalcPosition converts a byte offset to line and column (1-based).
func CalcPosition(source string, offset int) (line, column int) {
	if offset < 0 || offset > len(source) {
		return 1, 1
	}

	prefix := source[:offset]
	line = strings.Count(prefix, "\n") + 1
	lastNewline := strings.LastIndex(prefix, "\n")
	if lastNewline == -1 {
		column = offset + 1
	} else {
		column = offset - lastNewline
	}
	return line, column
}

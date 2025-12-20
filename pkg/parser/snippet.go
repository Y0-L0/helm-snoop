package parser

import (
	"log/slog"
	"text/template/parse"
)

// logNotImplementedCommand logs a concise message for not-implemented commands
// like include/tpl with file:line position and the command string.
func logNotImplementedCommand(tree *parse.Tree, ident string, cmd *parse.CommandNode) {
	if ident != "include" && ident != "tpl" {
		return
	}
	loc, _ := tree.ErrorContext(cmd)
	slog.Warn(ident+" not implemented", "cmd", cmd.String(), "pos", loc)
}

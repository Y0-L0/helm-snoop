package parser

import "log/slog"

func passthrough1Fn(args ...interface{}) interface{} {
	if len(args) == 0 {
		slog.Warn("Passthrough function called with no arguments", "args", args)
		Must("Passthrough function called with no arguments")
		return nil
	}
	return args[0]
}

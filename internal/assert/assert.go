package assert

import "log/slog"

// Strict controls whether internal assertions panic.
// Default true: tests run in strict mode. Executables disable it in main.
var Strict = true

// Must panics with msg when Strict is true; otherwise logs at Info level.
func Must(msg string) {
	if Strict {
		panic(msg)
	}
	slog.Info(msg)
}

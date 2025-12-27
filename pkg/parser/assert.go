package parser

// Strict controls whether internal assertions panic.
// Default true: tests run in strict mode. Executables disable it in main.
var Strict = true

// Must panics with msg when Strict is true; otherwise it does nothing.
func Must(msg string) {
	if Strict {
		panic(msg)
	}
}

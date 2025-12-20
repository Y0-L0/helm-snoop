package parser

// Strict controls whether internal assertions panic.
// Default true: tests run in strict mode. Executables disable it in main.
var Strict = true

// must panics with msg when Strict is true; otherwise it does nothing.
func must(msg string) {
	if Strict {
		panic(msg)
	}
}

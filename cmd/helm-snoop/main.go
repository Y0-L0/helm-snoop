package main

import (
	"os"

	"github.com/y0-l0/helm-snoop/pkg/parser"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
)

func main() {
	// Disable strict panics for production runs; tests keep the default (true).
	parser.Strict = false
	os.Exit(snooper.Main(os.Args, os.Stdout, os.Stderr))
}

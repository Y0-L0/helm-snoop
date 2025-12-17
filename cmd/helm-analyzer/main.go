package main

import (
	"os"

	"github.com/y0-l0/helm-analyzer/analyzer"
)

func main() { os.Exit(analyzer.Main(os.Args, os.Stdout, os.Stderr)) }

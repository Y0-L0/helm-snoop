package main

import (
	"os"

	"github.com/y0-l0/helm-snoop/pkg/snooper"
)

func main() {
	os.Exit(snooper.Main(os.Args, os.Stdout, os.Stderr))
}

package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/y0-l0/helm-snoop/pkg/parser"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

type LoaderFunc func(string) (*chart.Chart, error)

type cliConfig struct {
	chartPath  string
	ignoreKeys []string
	jsonOutput bool
	outWriter  io.Writer
	snoop      snooper.SnoopFunc
	loader     LoaderFunc
}

func (c *cliConfig) analyze() error {
	parser.Strict = false

	chart, err := c.loader(c.chartPath)
	if err != nil {
		fmt.Fprintf(c.outWriter, "Failed to read helm chart.\nerror: %v\n", err)
		return errors.New("")
	}

	result, err := c.snoop(chart, c.ignoreKeys)
	if err != nil {
		fmt.Fprintf(c.outWriter, "Failed to analyze the helm chart.\nerror: %v\n", err)
		return errors.New("")
	}

	if c.jsonOutput {
		if err := result.ToJSON(c.outWriter); err != nil {
			return errors.New("")
		}
	} else {
		if err := result.ToText(c.outWriter); err != nil {
			return errors.New("")
		}
	}

	if result.HasFindings() {
		return errors.New("")
	}
	return nil
}

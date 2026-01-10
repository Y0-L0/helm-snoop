package cli

import (
	"errors"
	"io"

	"github.com/y0-l0/helm-snoop/pkg/parser"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
)

type cliConfig struct {
	chartPath  string
	ignoreKeys []string
	jsonOutput bool
	outWriter  io.Writer
	snoop      snooper.SnoopFunc
}

func (c *cliConfig) analyze() error {
	parser.Strict = false

	result, err := c.snoop(c.chartPath, c.ignoreKeys)
	if err != nil {
		return err
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

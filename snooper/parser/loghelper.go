package parser

import (
	"log/slog"

	"github.com/davecgh/go-spew/spew"
)

type multiline string

func (m multiline) String() string { return string(m) }

type lazySpew struct{ v any }

func (l lazySpew) LogValue() slog.Value {
	cfg := spew.ConfigState{
		MaxDepth:                4,
		Indent:                  "  ",
		DisableMethods:          true,
		DisablePointerMethods:   true,
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
		SpewKeys:                true,
	}
	result := cfg.Sdump(l.v)
	return slog.AnyValue(multiline(result))
}

package parser

import (
	"fmt"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

var noopFuncs = []string{
	// All functions that use unaryPassThroughFn - they evaluate their
	// single argument and pass through .Values paths unchanged
	"lower",
	"mustToJson",
	"mustToYaml",
	"quote",
	"toJson",
	"toToml",
	"toYaml",
	"toYamlPretty",
	"upper",
}

func (s *Unittest) TestParseCommand_Noop() {
	expected := path.Paths{path.NewPath("app", "name")}

	for _, funcName := range noopFuncs {
		s.Run(funcName, func() {

			tmpl := fmt.Sprintf(`{{ %s .Values.app.name }}`, funcName)

			actual, err := parseFile(funcName+".tmpl", []byte(tmpl), nil)
			s.Require().NoError(err)

			path.EqualPaths(s, expected, actual)
		})
	}
}

func (s *Unittest) TestParseCommand_Noop_Pipe() {
	expected := path.Paths{path.NewPath("app", "name")}

	for _, funcName := range noopFuncs {
		s.Run(funcName, func() {

			tmpl := fmt.Sprintf(`{{ .Values.app.name | %s }}`, funcName)

			actual, err := parseFile(funcName+".tmpl", []byte(tmpl), nil)
			s.Require().NoError(err)

			path.EqualPaths(s, expected, actual)
		})
	}
}

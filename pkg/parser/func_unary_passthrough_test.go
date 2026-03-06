package parser

import (
	"fmt"

	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

var passthroughFuncs = []string{
	// Functions that use unaryPassThroughFn - they evaluate their
	// single argument and pass through .Values paths unchanged
	"lower",
	"quote",
	"upper",
}

var serializeFuncs = []string{
	// Functions that use unarySerializeFn - they serialize entire subtrees
	// and emit paths with a terminal wildcard
	"mustToJson",
	"mustToYaml",
	"toJson",
	"toToml",
	"toYaml",
	"toYamlPretty",
}

func (s *Unittest) TestParseCommand_Noop() {
	expected := vpath.Paths{vpath.NewPath("app", "name")}

	for _, funcName := range passthroughFuncs {
		s.Run(funcName, func() {
			tmpl := fmt.Sprintf(`{{ %s .Values.app.name }}`, funcName)
			vpath.EqualPaths(s, expected, s.parse(tmpl))
		})
	}
}

func (s *Unittest) TestParseCommand_Noop_Pipe() {
	expected := vpath.Paths{vpath.NewPath("app", "name")}

	for _, funcName := range passthroughFuncs {
		s.Run(funcName, func() {
			tmpl := fmt.Sprintf(`{{ .Values.app.name | %s }}`, funcName)
			vpath.EqualPaths(s, expected, s.parse(tmpl))
		})
	}
}

func (s *Unittest) TestParseCommand_Serialize() {
	expected := vpath.Paths{np().Key("app").Key("name").Wildcard()}

	for _, funcName := range serializeFuncs {
		s.Run(funcName, func() {
			tmpl := fmt.Sprintf(`{{ %s .Values.app.name }}`, funcName)
			vpath.EqualPaths(s, expected, s.parse(tmpl))
		})
	}
}

func (s *Unittest) TestParseCommand_Serialize_Pipe() {
	expected := vpath.Paths{np().Key("app").Key("name").Wildcard()}

	for _, funcName := range serializeFuncs {
		s.Run(funcName, func() {
			tmpl := fmt.Sprintf(`{{ .Values.app.name | %s }}`, funcName)
			vpath.EqualPaths(s, expected, s.parse(tmpl))
		})
	}
}

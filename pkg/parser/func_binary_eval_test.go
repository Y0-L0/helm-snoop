package parser

import (
	"fmt"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// binaryEvalFuncs are functions that take 2 arguments
// and should evaluate both to find .Values paths.
// They should use a pattern like emitArgsNoResultFn -
// evaluate all args and emit their paths.
//
// Reference: https://helm.sh/docs/chart_template_guide/function_list/

var binaryEvalFuncs = []string{
	// String testing/manipulation
	"contains",
	"eq",
	"ge",
	"gt",
	"has",
	"hasKey",
	"hasPrefix",
	"hasSuffix",
	"kindIs",
	"le",
	"lt",
	"ne",
	"regexFind",
	"regexMatch",
	"trimPrefix",
	"trimSuffix",
	"typeIs",
	"typeIsLike",
}

// TestParseCommand_BinaryEval tests functions that take 2 args (or 1 + piped)
// and should find .Values paths in both arguments
func (s *Unittest) TestParseCommand_BinaryEval() {
	testCases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name:     "values_in_first_arg",
			template: `{{ %s .Values.app.name "literal" }}`,
			expected: path.Paths{path.NewPath("app", "name")},
		},
		{
			name:     "values_in_second_arg",
			template: `{{ %s "literal" .Values.app.name }}`,
			expected: path.Paths{path.NewPath("app", "name")},
		},
		{
			name:     "values_in_both_args",
			template: `{{ %s .Values.app.name .Values.app.version }}`,
			expected: path.Paths{path.NewPath("app", "name"), path.NewPath("app", "version")},
		},
		{
			name:     "piped_form",
			template: `{{ .Values.app.name | %s "literal" }}`,
			expected: path.Paths{path.NewPath("app", "name")},
		},
	}

	for _, tc := range testCases {
		for _, funcName := range binaryEvalFuncs {
			testName := fmt.Sprintf("%s/%s", tc.name, funcName)
			s.Run(testName, func() {
				tmpl := fmt.Sprintf(tc.template, funcName)

				actual, err := parseFile("", funcName+".tmpl", []byte(tmpl), nil)
				s.Require().NoError(err)
				path.EqualPaths(s, tc.expected, actual)
			})
		}
	}
}

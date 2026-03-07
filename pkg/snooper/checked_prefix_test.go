package snooper

import (
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"

	"github.com/y0-l0/helm-snoop/pkg/tplparser"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

// {{ if .Values.parent }} should not be Undefined when parent.usedLeaf exists.
func (s *Unittest) TestCheckedPrefix_IfConditionOnParentObject() {
	restore := disableStrictParsing()
	defer restore()

	values := []byte(`
parent:
  usedLeaf: ""
  unusedLeaf: []
`)

	c := &chart.Chart{
		Metadata: &chart.Metadata{Name: "test"},
		Raw:      []*common.File{{Name: "values.yaml", Data: values}},
		Templates: []*common.File{
			{
				Name: "templates/_helpers.yaml",
				Data: []byte(`{{ define "lib.tpl" }}` +
					`{{ if .parent }}{{ .parent.usedLeaf }}{{ end }}` +
					`{{ end }}`),
			},
			{
				Name: "templates/main.yaml",
				Data: []byte(
					`{{ include "lib.tpl" (dict "parent" .Values.parent) }}` + "\n" +
						`{{ if .Values.absent }}x{{ end }}`),
			},
		},
	}

	used, err := tplparser.GetUsages(c)
	s.Require().NoError(err)

	defined, err := vpath.GetDefinitions(values, "values.yaml")
	s.Require().NoError(err)

	_, unused, undefined := vpath.MergeJoinLoose(defined, used)

	s.Equal([]string{".parent.unusedLeaf"}, pathIDs(unused))
	s.Equal([]string{".absent"}, pathIDs(undefined))
}

func pathIDs(paths vpath.Paths) []string {
	ids := make([]string, len(paths))
	for i, p := range paths {
		ids[i] = p.ID()
	}
	return ids
}

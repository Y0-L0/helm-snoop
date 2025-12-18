package path

import (
	"log/slog"
	"sort"

	"gopkg.in/yaml.v3"
)

func (s *Unittest) EqualPaths(expected Paths, actual Paths) {
	s.Require().Equal(len(expected), len(actual))
	sort.Sort(expected)
	sort.Sort(actual)

	for i, exp := range expected {
		s.Equal(exp, actual[i])
	}
}

func (s *Unittest) TestParseYaml() {
	testCases := []struct {
		name     string
		expected []*Path
		values   string
	}{
		{
			name: "simple map",
			expected: []*Path{
				np().Key("key1"),
				np().Key("key2"),
			},
			values: `
key1: value1
key2: value2
`,
		},
		{
			name: "complex",
			expected: []*Path{
				np().Key("list").Idx("0"),
				np().Key("list").Idx("1"),
				np().Key("nestedMap").Key("nestedKey"),
				np().Key("nestedMap").Key("nestedList").Idx("0"),
				np().Key("nestedMap").Key("nestedList").Idx("1"),
			},
			values: `
list:
  - item1
  - item2
nestedMap:
  nestedKey: nestedValue
  nestedList:
    - nestedItem1
    - nestedItem2
`,
		},
		{
			name: "list of maps",
			expected: []*Path{
				np().Idx("0").Key("listKey1").Key("value1").Key("nestedKey"),
				np().Idx("1").Key("listKey2"),
			},
			values: `
- listKey1:
    value1:
      nestedKey: nestedValue
- listKey2: value2
`,
		},
		{
			name:     "integer key",
			expected: []*Path{np().Key("1234").Key("integerValue")},
			values: `
1234:
  integerValue: something
`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var values interface{}
			err := yaml.Unmarshal([]byte(tc.values), &values)
			s.Require().NoError(err)
			slog.Debug("complete test values.yaml", "yaml", values)

			out := []*Path{}
			GetDefinitions(Path{}, values, &out)

			s.EqualPaths(Paths(tc.expected), Paths(out))
		})
	}
}

func (s *Unittest) TestFlattenValues_NonStringKeys() {
	values := map[interface{}]interface{}{
		"stringKey": "value1",
		123:         "value2", // non-string key
	}

	out := []*Path{}
	GetDefinitions(Path{}, values, &out)

	expected := []*Path{
		np().Key("stringKey"),
		np().Key("123"),
	}

	s.EqualPaths(Paths(expected), Paths(out))
}

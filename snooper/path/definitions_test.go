package path

import (
	"log/slog"

	"gopkg.in/yaml.v3"
)

func (s *Unittest) TestParseYaml() {
	testCases := []struct {
		name     string
		expected Paths
		values   string
	}{
		{
			name: "simple map",
			expected: Paths{
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
			expected: Paths{
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
			expected: Paths{
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
			expected: Paths{np().Key("1234").Key("integerValue")},
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

			out := Paths{}
			GetDefinitions(Path{}, values, &out)

			EqualPaths(s, Paths(tc.expected), Paths(out))
		})
	}
}

func (s *Unittest) TestFlattenValues_NonStringKeys() {
	values := map[interface{}]interface{}{
		"stringKey": "value1",
		123:         "value2", // non-string key
	}

	out := Paths{}
	GetDefinitions(Path{}, values, &out)

	expected := Paths{
		np().Key("stringKey"),
		np().Key("123"),
	}

	EqualPaths(s, Paths(expected), Paths(out))
}

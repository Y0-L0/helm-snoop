package vpath

func (s *Unittest) TestGetDefinitions() {
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
			out, err := GetDefinitions([]byte(tc.values), "values.yaml")
			s.Require().NoError(err)

			EqualPaths(s, tc.expected, out)
		})
	}
}

func (s *Unittest) TestGetDefinitions_WithContext() {
	values := "config:\n  database:\n    host: localhost\nitems:\n  - name: foo\n"

	out, err := GetDefinitions([]byte(values), "values.yaml")
	s.Require().NoError(err)

	byID := map[string]Context{}
	for _, p := range out {
		s.Require().Len(p.Contexts, 1)
		byID[p.ID()] = p.Contexts[0]
	}

	s.Equal(Context{FileName: "values.yaml", Line: 3, Column: 5},
		byID[".config.database.host"])
	s.Equal(Context{FileName: "values.yaml", Line: 5, Column: 5},
		byID[".items.0.name"])
}

func (s *Unittest) TestGetDefinitions_EmptyCollections() {
	testCases := []struct {
		name     string
		expected Paths
		values   string
	}{
		{
			name: "empty_dict",
			expected: Paths{
				np().Key("podAnnotations"),
			},
			values: "podAnnotations: {}\n",
		},
		{
			name: "empty_list",
			expected: Paths{
				np().Key("items"),
			},
			values: "items: []\n",
		},
		{
			name: "nested_empty_collections",
			expected: Paths{
				np().Key("config").Key("annotations"),
				np().Key("config").Key("labels"),
			},
			values: `
config:
  annotations: {}
  labels: {}
`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := GetDefinitions([]byte(tc.values), "values.yaml")
			s.Require().NoError(err)

			EqualPaths(s, tc.expected, out)
		})
	}
}

func (s *Unittest) TestGetDefinitions_EmptyCollectionContexts() {
	values := "podAnnotations: {}\nitems: []\n"

	out, err := GetDefinitions([]byte(values), "values.yaml")
	s.Require().NoError(err)

	byID := map[string]Context{}
	for _, p := range out {
		s.Require().Len(p.Contexts, 1)
		byID[p.ID()] = p.Contexts[0]
	}

	s.Equal(Context{FileName: "values.yaml", Line: 1, Column: 1},
		byID[".podAnnotations"])
	s.Equal(Context{FileName: "values.yaml", Line: 2, Column: 1},
		byID[".items"])
}

func (s *Unittest) TestGetDefinitionsFromMap() {
	testCases := []struct {
		name     string
		m        map[string]any
		expected Paths
	}{
		{
			name: "nested map",
			m: map[string]any{
				"image": map[string]any{
					"repository": "nginx",
					"tag":        "latest",
				},
			},
			expected: Paths{
				np().Key("image").Key("repository"),
				np().Key("image").Key("tag"),
			},
		},
		{
			name:     "nil map",
			m:        nil,
			expected: nil,
		},
		{
			name:     "empty map",
			m:        map[string]any{},
			expected: nil,
		},
		{
			name: "scalar values",
			m: map[string]any{
				"replicas": 3,
				"enabled":  true,
				"name":     "test",
			},
			expected: Paths{
				np().Key("enabled"),
				np().Key("name"),
				np().Key("replicas"),
			},
		},
		{
			name: "sequence",
			m: map[string]any{
				"items": []any{"a", "b"},
			},
			expected: Paths{
				np().Key("items").Idx("0"),
				np().Key("items").Idx("1"),
			},
		},
		{
			name: "empty nested collections",
			m: map[string]any{
				"annotations": map[string]any{},
				"items":       []any{},
			},
			expected: Paths{
				np().Key("annotations"),
				np().Key("items"),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out := GetDefinitionsFromMap(tc.m, "test-source")
			EqualPaths(s, tc.expected, out)
		})
	}
}

func (s *Unittest) TestGetDefinitionsFromMap_HasSourceContext() {
	m := map[string]any{
		"key": "value",
	}
	out := GetDefinitionsFromMap(m, "my-config.yaml:global.extraValues")
	s.Require().Len(out, 1)
	s.Require().Len(out[0].Contexts, 1)
	s.Equal("my-config.yaml:global.extraValues", out[0].Contexts[0].FileName)
}

func (s *Unittest) TestFormatMapSource() {
	s.Equal(
		"config.yaml:global.extraValues",
		FormatMapSource("config.yaml", ""),
	)
	s.Equal(
		"config.yaml:charts.my-chart.extraValues",
		FormatMapSource("config.yaml", "my-chart"),
	)
}

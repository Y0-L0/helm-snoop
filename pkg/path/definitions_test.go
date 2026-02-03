package path

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

			EqualPaths(s, Paths(tc.expected), Paths(out))
		})
	}
}

func (s *Unittest) TestGetDefinitions_WithContext() {
	values := "config:\n  database:\n    host: localhost\nitems:\n  - name: foo\n"

	out, err := GetDefinitions([]byte(values), "values.yaml")
	s.Require().NoError(err)

	byID := map[string]PathContext{}
	for _, p := range out {
		s.Require().Len(p.Contexts, 1)
		byID[p.ID()] = p.Contexts[0]
	}

	s.Equal(PathContext{FileName: "values.yaml", Line: 3, Column: 5},
		byID[".config.database.host"])
	s.Equal(PathContext{FileName: "values.yaml", Line: 5, Column: 5},
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

			EqualPaths(s, Paths(tc.expected), Paths(out))
		})
	}
}

func (s *Unittest) TestGetDefinitions_EmptyCollectionContexts() {
	values := "podAnnotations: {}\nitems: []\n"

	out, err := GetDefinitions([]byte(values), "values.yaml")
	s.Require().NoError(err)

	byID := map[string]PathContext{}
	for _, p := range out {
		s.Require().Len(p.Contexts, 1)
		byID[p.ID()] = p.Contexts[0]
	}

	s.Equal(PathContext{FileName: "values.yaml", Line: 1, Column: 1},
		byID[".podAnnotations"])
	s.Equal(PathContext{FileName: "values.yaml", Line: 2, Column: 1},
		byID[".items"])
}

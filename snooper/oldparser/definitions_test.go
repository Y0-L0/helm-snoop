package oldparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestFlattenValues(t *testing.T) {
	testCases := []struct {
		name     string
		prefix   string
		expected []string
		values   string
	}{
		{
			name:     "simple map",
			prefix:   "",
			expected: []string{"key1", "key2"},
			values: `
key1: value1
key2: value2
`,
		},
		{
			name:     "complex",
			prefix:   "root",
			expected: []string{"root.list", "root.nestedMap.nestedKey", "root.nestedMap.nestedList"},
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
			name:     "list of maps",
			prefix:   "items",
			expected: []string{"items"},
			values: `
- listKey1:
    value1:
      nestedKey: nestedValue
- listKey2: value2
`,
		},
		{
			name:     "integer key",
			prefix:   "",
			expected: []string{"1234.integerValue"},
			values: `
1234:
  integerValue: something
`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var values interface{}
			err := yaml.Unmarshal([]byte(testCase.values), &values)
			assert.NoError(t, err)

			out := make(map[string]struct{})
			GetDefinitions(testCase.prefix, values, out)

			expectedMap := make(map[string]struct{})
			for _, item := range testCase.expected {
				expectedMap[item] = struct{}{}
			}

			assert.Equal(t, expectedMap, out)
		})
	}
}

func TestFlattenValues_NonStringKeys(t *testing.T) {
	t.Run("non-string key", func(t *testing.T) {
		values := map[interface{}]interface{}{
			"stringKey": "value1",
			123:         "value2", // non-string key
		}

		out := make(map[string]struct{})
		GetDefinitions("", values, out)

		expected := map[string]struct{}{
			"stringKey": {},
			"123":       {},
		}

		assert.Equal(t, expected, out)
	})
}

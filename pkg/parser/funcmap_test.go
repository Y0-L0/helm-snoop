package parser

import (
	"sort"

	"github.com/Masterminds/sprig/v3"
)

// Ensure all entries in evalRegistry are templFunc-typed so getTemplateFunction
// can retrieve them without signature panics.
func (s *Unittest) TestFuncMap_AllEntriesAreTemplFunc() {
	for k, v := range funcMap {
		if v == nil {
			s.T().Errorf("evalRegistry entry %q is nil", k)
		}
	}
}

// Compare our func map keys with sprig + Helm extras.
// This ensures we don't introduce unknown function names inadvertently.
// expectedFuncKeys builds the canonical key set by mirroring Helm's funcMap construction:
// sprig.TxtFuncMap() - {"env", "expandenv"} + Helm extras
func expectedFuncKeys() []string {
	// Start with sprig's text template functions
	f := sprig.TxtFuncMap()

	// Helm removes these for security
	delete(f, "env")
	delete(f, "expandenv")

	// Helm adds these extras (from helm/pkg/engine/funcs.go)
	extras := []string{
		"toToml", "fromToml",
		"toYaml", "mustToYaml", "toYamlPretty",
		"fromYaml", "fromYamlArray",
		"toJson", "mustToJson",
		"fromJson", "fromJsonArray",
		"include", "tpl", "required", "lookup",
	}

	// Helm also adds these in initFunMap (helm/pkg/engine/engine.go)
	helmExtras := []string{
		"fail", "getHostByName",
	}

	// Go template built-ins that appear as functions in templates
	// (from text/template/parse/parse.go builtins map)
	goTemplateBuiltins := []string{
		// Comparison functions
		"eq", "ne", "lt", "le", "gt", "ge",
		// Logical operators
		"and", "or", "not",
		// Utility functions
		"index", "len", "call",
		// Output functions
		"print", "printf", "println",
		// Escaping functions
		"html", "js", "urlquery",
	}

	// Add extras (may override existing Sprig keys)
	for _, k := range extras {
		f[k] = nil // Just mark as present
	}
	for _, k := range helmExtras {
		f[k] = nil
	}
	for _, k := range goTemplateBuiltins {
		f[k] = nil
	}

	// Extract unique keys from map
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

func (s *Unittest) TestFuncMap_KeysMatchExpected() {
	expected := expectedFuncKeys()
	actual := make([]string, 0, len(funcMap))
	for k := range funcMap {
		actual = append(actual, k)
	}
	sort.Strings(actual)

	s.Equal(expected, actual, "funcMap keys must match expected Helm + Sprig + Go template functions")
}

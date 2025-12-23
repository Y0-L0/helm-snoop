package parser

import "sort"

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
// expectedFuncKeys builds the canonical key set mirroring Helm's funcMap construction.
func expectedFuncKeys() []string {
	// Expected keys come from our own func map definition (no sprig).
	keys := []string{
		"include", "tpl", "default", "get", "index", "int", "concat",
		"quote", "upper", "lower", "indent", "nindent",
		"toYaml", "mustToYaml", "toYamlPretty",
		"toJson", "mustToJson", "toToml",
		"fromYaml", "fromYamlArray", "fromJson", "fromJsonArray",
		"dict", "merge", "mergeOverwrite", "omit", "contains", "replace", "slice",
		"sha256sum", "sha1sum", "semverCompare", "trimSuffix", "trunc", "urlquery",
		"print", "printf", "println", "trim",
		"or", "and", "not", "eq", "ne", "ge", "gt", "le", "lt", "len", "html", "js", "call", "pick", "kindIs",
		"fail", "lookup", "getHostByName", "required", "derivePassword", "empty", "has", "hasKey", "regexFind", "regexMatch", "regexReplaceAllLiteral", "substr", "toString", "typeIs", "typeOf", "typeIsLike", "b64dec", "b64enc", "list", "randAlpha", "randAlphaNum", "randNumeric", "join", "randAscii", "shuffle", "splitList", "keys", "reverse", "split", "uniq", "semver",
		"coalesce", "append", "first", "set", "ternary", "cat",
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
	s.Equal(len(expected), len(actual), "funcmap key count mismatch")
	s.Equal(expected, actual, "funcmap keys differ from expected")
}

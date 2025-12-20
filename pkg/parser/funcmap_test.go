package parser

// Ensure all entries in templFuncMap are templFunc-typed so getTemplateFunction
// can retrieve them without signature panics.
func (s *Unittest) TestFuncMap_AllEntriesAreTemplFunc() {
	for k, v := range templFuncMap {
		if _, ok := v.(templFunc); !ok {
			s.T().Errorf("funcmap entry %q is not templFunc-typed: %T", k, v)
		}
	}
}

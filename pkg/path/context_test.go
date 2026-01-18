package path

func (s *Unittest) TestPathContext_String() {
	ctx := PathContext{FileName: "templates/deployment.yaml", Line: 42, Column: 10}
	s.Equal("templates/deployment.yaml:42:10", ctx.String())
}

func (s *Unittest) TestPathContext_String_LineOne() {
	ctx := PathContext{FileName: "test.yaml", Line: 1, Column: 1}
	s.Equal("test.yaml:1:1", ctx.String())
}

func (s *Unittest) TestPathContext_StringWithTemplate() {
	ctx := PathContext{
		FileName:     "templates/_helpers.tpl",
		TemplateName: "mychart.fullname",
		Line:         15,
		Column:       5,
	}
	s.Equal("templates/_helpers.tpl:15:5 (mychart.fullname)", ctx.String())
}

func (s *Unittest) TestPathContext_StringEmptyTemplate() {
	ctx := PathContext{
		FileName:     "templates/deployment.yaml",
		TemplateName: "",
		Line:         42,
		Column:       10,
	}
	s.Equal("templates/deployment.yaml:42:10", ctx.String())
}

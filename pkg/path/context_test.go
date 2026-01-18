package path

import "encoding/json"

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

func (s *Unittest) TestPath_ToJSON_Integration() {
	p := NewPath("config", "database", "host")
	p.Contexts = []PathContext{
		{FileName: "templates/deployment.yaml", Line: 42, Column: 10},
		{FileName: "templates/service.yaml", TemplateName: "myhelper", Line: 5, Column: 3},
	}

	bytes, err := json.Marshal(p.ToJSON())
	s.Require().NoError(err)

	expected := `{"id":"/config/database/host","kinds":"/K/K/K","contexts":[{"file":"templates/deployment.yaml","line":42,"column":10},{"file":"templates/service.yaml","template":"myhelper","line":5,"column":3}]}`
	s.Equal(expected, string(bytes))
}

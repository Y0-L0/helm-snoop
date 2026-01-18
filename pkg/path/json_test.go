package path

func (s *Unittest) TestPathJSON_WithContext() {
	p := NewPath("config", "name")
	p.Contexts = []PathContext{
		{FileName: "templates/deployment.yaml", Line: 42, Column: 10},
	}

	j := p.ToJSON()

	s.Equal("/config/name", j.ID)
	s.Equal("/K/K", j.Kinds)
	s.Require().Len(j.Contexts, 1)
	s.Equal("templates/deployment.yaml", j.Contexts[0].File)
	s.Equal(42, j.Contexts[0].Line)
	s.Equal(10, j.Contexts[0].Column)
}

func (s *Unittest) TestPathJSON_WithTemplateContext() {
	p := NewPath("foo")
	p.Contexts = []PathContext{
		{FileName: "templates/_helpers.tpl", TemplateName: "mychart.name", Line: 5, Column: 3},
	}

	j := p.ToJSON()

	s.Require().Len(j.Contexts, 1)
	s.Equal("mychart.name", j.Contexts[0].Template)
}

func (s *Unittest) TestPathJSON_NoContext() {
	p := NewPath("foo")
	j := p.ToJSON()

	s.Nil(j.Contexts)
}

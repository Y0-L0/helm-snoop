package tplparser

// Two files, one define each → both indexed with correct origins.
func (s *Unittest) TestBuildTemplateIndex_TwoFilesTwoDefines() {
	c := buildChart(
		testFile{"templates/a.yaml", `{{ define "tpl.one" }}{{ .Values.a }}{{ end }}`},
		testFile{"templates/b.yaml", `{{ define "tpl.two" }}{{ .Values.b }}{{ end }}`},
	)

	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)

	def1, ok1 := idx.get("tpl.one")
	def2, ok2 := idx.get("tpl.two")
	s.Require().True(ok1)
	s.Require().True(ok2)
	s.Require().Equal("templates/a.yaml", def1.file)
	s.Require().Equal("templates/b.yaml", def2.file)
	s.Require().NotNil(def1.root)
	s.Require().NotNil(def2.root)
}

// Multiple defines in the same file → all indexed to that file.
func (s *Unittest) TestBuildTemplateIndex_SameFileTwoDefines() {
	c := buildChart(testFile{"templates/c.yaml", `{{ define "tpl.one" }}x{{ end }}{{ define "tpl.two" }}y{{ end }}`})
	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)

	def1, ok1 := idx.get("tpl.one")
	def2, ok2 := idx.get("tpl.two")
	s.Require().True(ok1)
	s.Require().True(ok2)
	s.Require().Equal("templates/c.yaml", def1.file)
	s.Require().Equal("templates/c.yaml", def2.file)
}

// No define blocks → empty index.
func (s *Unittest) TestBuildTemplateIndex_NoDefines() {
	c := buildChart(testFile{"templates/d.yaml", `kind: ConfigMap`})
	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)
	s.Require().True(idx.empty())
}

// Duplicate define names across files → panic (Strict in tests).
func (s *Unittest) TestBuildTemplateIndex_DuplicateNames() {
	c := buildChart(
		testFile{"templates/a.yaml", `{{ define "tpl.one" }}x{{ end }}`},
		testFile{"templates/b.yaml", `{{ define "tpl.one" }}y{{ end }}`},
	)
	s.Require().Panics(func() { _, _ = BuildTemplateIndex(c) })
}

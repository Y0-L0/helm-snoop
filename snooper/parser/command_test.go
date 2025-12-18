package parser

// quote/upper/lower simply wrap a value; default X Y reads only Y when X is literal.
func (s *Unittest) TestParseCommand_Noops() {
	cases := []struct {
		name string
		tmpl []string
		want []string
	}{
		{
			name: "quote",
			tmpl: []string{
				`{{ quote .Values.app.name }}`,
				`{{ .Values.app.name | quote }}`,
			},
			want: []string{"app.name"},
		},
		{
			name: "upper",
			tmpl: []string{
				`{{ upper .Values.ns }}`,
				`{{ .Values.ns | upper }}`,
			},
			want: []string{"ns"},
		},
		{
			name: "lower",
			tmpl: []string{
				`{{ lower .Values.Kind }}`,
				`{{ .Values.Kind | lower }}`,
			},
			want: []string{"Kind"},
		},
		{
			name: "default literal + value",
			tmpl: []string{
				`{{ default "x" .Values.cfg.path }}`,
				`{{ "x" | default .Values.cfg.path }}`,
			},
			want: []string{"cfg.path"},
		},
	}

	for _, tc := range cases {
		for i, tmpl := range tc.tmpl {
			name := tc.name
			if i == 1 {
				name += "_piped"
			}
			s.Run(name, func() {
				got, err := parseFile(name+".tmpl", []byte(tmpl))
				s.Require().NoError(err)
				s.Require().Equal(tc.want, got)
			})
		}
	}
}

// get/index merge their arguments and return them.
func (s *Unittest) TestParseCommand_Return() {
	cases := []struct {
		name string
		tmpl []string
		want []string
	}{
		{
			name: "get wrapper",
			tmpl: []string{
				`{{ get .Values.app "name"}}`,
				`{{ "name" | get .Values.app }}`,
			},
			want: []string{"app.name"},
		},
		{
			name: "index one",
			tmpl: []string{
				`{{ index .Values.cfg.path "firstIndex" }}`,
				`{{ "firstIndex" | index index .Values.cfg.path }}`,
			},
			want: []string{"cfg.path.firstIndex"},
		},
		{
			name: "index two",
			tmpl: []string{
				`{{ index .Values.cfg.path "firstIndex" "secondIndex" }}`,
				`{{ "secondIndex" | index .Values.cfg.path "firstIndex" }}`,
			},
			want: []string{"cfg.path.firstIndex.secondIndex"},
		},
	}

	for _, tc := range cases {
		for i, tmpl := range tc.tmpl {
			name := tc.name
			if i == 1 {
				name += "_piped"
			}
			s.Run(name, func() {
				got, err := parseFile(tc.name+".tmpl", []byte(tmpl))
				s.Require().NoError(err)
				s.Require().Equal(tc.want, got)
			})
		}
	}
}

// Functions out of scope: include/tpl should be treated as not implemented.
func (s *Unittest) TestParseCommand_NotImplemented() {
	cases := []struct {
		name string
		tmpl string
	}{
		{name: "include function", tmpl: `{{ include "x" . }}`},
		{name: "tpl function", tmpl: `{{ tpl "{{ .Values.a.x }}" . }}`},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.Require().Panics(func() {
				_, _ = parseFile(tc.name+".tmpl", []byte(tc.tmpl))
			})
		})
	}
}

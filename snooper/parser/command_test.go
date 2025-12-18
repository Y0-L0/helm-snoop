package parser

// Functions that should behave as no-ops for extraction in Stage 2.
// quote/upper/lower simply wrap a value; default X Y reads only Y when X is literal.
func (s *Unittest) TestParseCommand_Noops() {
	cases := []struct {
		name string
		tmpl string
		want []string
	}{
		{
			name: "quote wrapper",
			tmpl: `{{ quote .Values.app.name }}`,
			want: []string{"app.name"},
		},
		{
			name: "upper wrapper",
			tmpl: `{{ upper .Values.ns }}`,
			want: []string{"ns"},
		},
		{
			name: "lower wrapper",
			tmpl: `{{ lower .Values.Kind }}`,
			want: []string{"Kind"},
		},
		{
			name: "default literal + value",
			tmpl: `{{ default "x" .Values.cfg.path }}`,
			want: []string{"cfg.path"},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			got, err := parseFile(tc.name+".tmpl", []byte(tc.tmpl))
			s.Require().NoError(err)
			s.Require().Equal(tc.want, got)
		})
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

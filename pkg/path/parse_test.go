package path

func (s *Unittest) TestParsePath_ValidPaths() {
	tests := []struct {
		name  string
		input string
		want  *Path
	}{
		// Valid exact paths — with leading dot (matches ID() output)
		{
			name:  "simple key path with leading dot",
			input: ".image.tag",
			want:  np().Key("image").Key("tag"),
		},
		{
			name:  "single segment with leading dot",
			input: ".replicas",
			want:  np().Key("replicas"),
		},
		{
			name:  "deep nested path with leading dot",
			input: ".config.nested.deep.value",
			want:  np().Key("config").Key("nested").Key("deep").Key("value"),
		},

		// Without leading dot (also accepted)
		{
			name:  "simple key path without leading dot",
			input: "image.tag",
			want:  np().Key("image").Key("tag"),
		},
		{
			name:  "single segment without leading dot",
			input: "replicas",
			want:  np().Key("replicas"),
		},

		// Integer handling (anyKind)
		{
			name:  "integer as anyKind",
			input: ".items.0",
			want:  np().Key("items").Any("0"),
		},
		{
			name:  "multiple integers",
			input: ".config.1.nested.2",
			want:  np().Key("config").Any("1").Key("nested").Any("2"),
		},
		{
			name:  "large integer",
			input: ".items.12345",
			want:  np().Key("items").Any("12345"),
		},

		// Wildcard handling
		{
			name:  "terminal wildcard",
			input: ".config.*",
			want:  np().Key("config").Wildcard(),
		},
		{
			name:  "interior wildcard",
			input: ".a.*.c",
			want:  np().Key("a").Wildcard().Key("c"),
		},
		{
			name:  "multiple interior wildcards",
			input: ".a.*.c.*.e",
			want:  np().Key("a").Wildcard().Key("c").Wildcard().Key("e"),
		},
		{
			name:  "wildcard with integer",
			input: ".items.*.value",
			want:  np().Key("items").Wildcard().Key("value"),
		},

		// Mixed patterns
		{
			name:  "key then integer then wildcard",
			input: ".config.0.*",
			want:  np().Key("config").Any("0").Wildcard(),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := ParsePath(tc.input)
			s.Require().NoError(err)
			s.Require().Equal(tc.want, got)
		})
	}
}

func (s *Unittest) TestParsePath_ErrorCases() {
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "slash notation rejected",
			input:       "/image/tag",
			errContains: "dot notation",
		},
		{
			name:        "empty pattern",
			input:       "",
			errContains: "empty pattern",
		},
		{
			name:        "only dot",
			input:       ".",
			errContains: "empty pattern",
		},
		{
			name:        "double dot",
			input:       ".config..value",
			errContains: "empty segment",
		},
		{
			name:        "trailing dot",
			input:       ".config.",
			errContains: "trailing dot",
		},
		{
			name:        "multiple dots",
			input:       "...",
			errContains: "empty segment",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, err := ParsePath(tc.input)
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.errContains)
		})
	}
}

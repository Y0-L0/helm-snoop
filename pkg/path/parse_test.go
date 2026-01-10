package path

func (s *Unittest) TestParsePath_ValidPaths() {
	tests := []struct {
		name  string
		input string
		want  *Path
	}{
		// Valid exact paths
		{
			name:  "simple key path",
			input: "/image/tag",
			want:  np().Key("image").Key("tag"),
		},
		{
			name:  "single segment",
			input: "/replicas",
			want:  np().Key("replicas"),
		},
		{
			name:  "deep nested path",
			input: "/config/nested/deep/value",
			want:  np().Key("config").Key("nested").Key("deep").Key("value"),
		},

		// Integer handling (anyKind)
		{
			name:  "integer as anyKind",
			input: "/items/0",
			want:  np().Key("items").Any("0"),
		},
		{
			name:  "multiple integers",
			input: "/config/1/nested/2",
			want:  np().Key("config").Any("1").Key("nested").Any("2"),
		},
		{
			name:  "large integer",
			input: "/items/12345",
			want:  np().Key("items").Any("12345"),
		},

		// Wildcard handling
		{
			name:  "terminal wildcard",
			input: "/config/*",
			want:  np().Key("config").Wildcard(),
		},
		{
			name:  "interior wildcard",
			input: "/a/*/c",
			want:  np().Key("a").Wildcard().Key("c"),
		},
		{
			name:  "multiple interior wildcards",
			input: "/a/*/c/*/e",
			want:  np().Key("a").Wildcard().Key("c").Wildcard().Key("e"),
		},
		{
			name:  "wildcard with integer",
			input: "/items/*/value",
			want:  np().Key("items").Wildcard().Key("value"),
		},

		// Mixed patterns
		{
			name:  "key then integer then wildcard",
			input: "/config/0/*",
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
			name:        "no leading slash",
			input:       "image/tag",
			errContains: "must start with /",
		},
		{
			name:        "empty pattern",
			input:       "",
			errContains: "empty pattern",
		},
		{
			name:        "only slash",
			input:       "/",
			errContains: "empty pattern",
		},
		{
			name:        "double slash",
			input:       "/config//value",
			errContains: "empty segment",
		},
		{
			name:        "trailing slash",
			input:       "/config/",
			errContains: "trailing slash",
		},
		{
			name:        "multiple slashes",
			input:       "///",
			errContains: "trailing slash",
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

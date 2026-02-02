package path

func (s *Unittest) TestPathIDAndKinds() {
	cases := []struct {
		name          string
		path          *Path
		expectedID    string
		expectedKinds string
	}{
		{
			name:          "empty_path",
			path:          np(),
			expectedID:    ".",
			expectedKinds: "",
		},
		{
			name:          "single_wildcard",
			path:          np().Key("a").Wildcard(),
			expectedID:    ".a.*",
			expectedKinds: "/K/W",
		},
		{
			name:          "wildcard_at_end",
			path:          np().Key("config").Key("nested").Wildcard(),
			expectedID:    ".config.nested.*",
			expectedKinds: "/K/K/W",
		},
		{
			name:          "wildcard_in_middle",
			path:          np().Key("a").Wildcard().Key("b"),
			expectedID:    ".a.*.b",
			expectedKinds: "/K/W/K",
		},
		{
			name:          "key_with_dot_escaped",
			path:          np().Key("foo.bar").Key("baz"),
			expectedID:    ".foo~.bar.baz",
			expectedKinds: "/K/K",
		},
		{
			name:          "key_with_tilde_escaped",
			path:          np().Key("foo~bar"),
			expectedID:    ".foo~~bar",
			expectedKinds: "/K",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.Equal(tc.expectedID, tc.path.ID())
			s.Equal(tc.expectedKinds, tc.path.KindsString())
		})
	}
}

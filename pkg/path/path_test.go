// TestWildcardKind tests creating paths with wildcard segments
package path

func (s *Unittest) TestWildcardKind() {
	cases := []struct {
		name          string
		path          *Path
		expectedID    string
		expectedKinds string
	}{
		{
			name:          "single_wildcard",
			path:          np().Key("a").Wildcard(),
			expectedID:    "/a/*",
			expectedKinds: "/K/W",
		},
		{
			name:          "wildcard_at_end",
			path:          np().Key("config").Key("nested").Wildcard(),
			expectedID:    "/config/nested/*",
			expectedKinds: "/K/K/W",
		},
		{
			name:          "wildcard_in_middle",
			path:          np().Key("a").Wildcard().Key("b"),
			expectedID:    "/a/*/b",
			expectedKinds: "/K/W/K",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.Equal(tc.expectedID, tc.path.ID())
			s.Equal(tc.expectedKinds, tc.path.KindsString())
		})
	}
}

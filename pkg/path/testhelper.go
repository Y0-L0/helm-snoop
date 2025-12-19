package path

import (
	"log/slog"
	"sort"
)

type equality interface {
	Equal(interface{}, interface{}, ...interface{}) bool
}

func getPath(paths Paths, i int) *Path {
	if i < len(paths) {
		return paths[i]
	}
	return &Path{}
}

func EqualPaths(eq equality, expected Paths, actual Paths) {
	slog.Debug("Asserting equal Paths", "expected", expected, "actual", actual)
	sort.Sort(expected)
	sort.Sort(actual)
	EqualInorderPaths(eq, expected, actual)
}

func EqualInorderPaths(eq equality, expected Paths, actual Paths) {
	slog.Debug("Asserting equal in-order Paths", "expected", expected, "actual", actual)
	eq.Equal(len(expected), len(actual))
	for i := 0; i < len(expected) || i < len(actual); i++ {
		act := getPath(actual, i)
		exp := getPath(expected, i)
		eq.Equal(exp, act)
	}
}

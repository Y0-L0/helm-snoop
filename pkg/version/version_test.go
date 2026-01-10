package version

import (
	"bytes"
)

func (s *Unittest) TestBuildInfo() {
	var buf bytes.Buffer
	BuildInfo(&buf)

	output := buf.String()

	expectedFields := []string{"Version:", "Commit:", "TreeState:", "BuildDate:", "GoVersion:", "Platform:"}
	for _, field := range expectedFields {
		s.Contains(output, field)
	}

	s.Contains(output, "Version:    dev")
}

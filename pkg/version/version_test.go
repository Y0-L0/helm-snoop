package version

import (
	"bytes"
)

func (s *Unittest) TestBuildInfo() {
	var buf bytes.Buffer
	BuildInfo(&buf)

	output := buf.String()

	expectedFields := []string{"Version:", "Commit:", "TreeState:", "CommitDate:", "GoVersion:", "Platform:"}
	for _, field := range expectedFields {
		s.Contains(output, field)
	}

	// When running via go test, ReadBuildInfo sets Main.Version to "(devel)".
	s.Contains(output, "Version:    (devel)")
}

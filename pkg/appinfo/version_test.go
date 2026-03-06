package appinfo

import (
	"bytes"
)

func (s *Unittest) TestWrite() {
	var buf bytes.Buffer
	Write(&buf)

	output := buf.String()

	expectedFields := []string{"Version:", "Commit:", "TreeState:", "CommitDate:", "GoVersion:", "Platform:"}
	for _, field := range expectedFields {
		s.Contains(output, field)
	}

	// When running via go test, ReadBuildInfo sets Main.Version to "(devel)".
	s.Contains(output, "Version:    (devel)")
}

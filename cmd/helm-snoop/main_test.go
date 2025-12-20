package main

import (
	"bytes"
	"os"
)

func (s *GoldenTest) TestCLI_Usage() {
	restore, _, err, code := interceptMain([]string{"helm-snoop"})
	defer restore()
	main()
	s.Require().Equal(2, *code)
	s.EqualGoldenFile("usage.golden.txt", err.Bytes())
}

func (s *GoldenTest) TestCLI_Missing() {
	restore, _, err, code := interceptMain([]string{"helm-snoop", "does-not-exist"})
	defer restore()
	main()
	s.Require().Equal(1, *code)
	s.EqualGoldenFile("missing.golden.txt", err.Bytes())
}

func (s *GoldenTest) TestCLI_TestChart() {
	restore, stdout, stderr, code := interceptMain([]string{"helm-snoop", s.chartPath, "debug"})
	defer restore()
	main()
	s.Require().Equal(0, *code, stderr.String())
	s.EqualGoldenFile("test-chart.golden.txt", stdout.Bytes())
}

func interceptMain(args []string) (restore func(), out *bytes.Buffer, errBuf *bytes.Buffer, code *int) {
	oldExit, oldOut, oldErr := osExit, stdout, stderr
	outBuf, errB := &bytes.Buffer{}, &bytes.Buffer{}
	stdout, stderr = outBuf, errB

	var exitCode int
	osExit = func(c int) { exitCode = c }
	os.Args = args

	return func() { osExit, stdout, stderr = oldExit, oldOut, oldErr }, outBuf, errB, &exitCode
}

func (s *GoldenTest) EqualGoldenFile(golfenFileName string, actual []byte) {
	goldenFilePath := s.goldenFile(golfenFileName)
	if s.update {
		s.writeFile(goldenFilePath, actual)
	}
	expected := s.readFile(goldenFilePath)
	exp, act := string(expected), string(actual)
	s.Require().Equal(exp, act)
}

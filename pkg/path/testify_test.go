package path

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/y0-l0/helm-snoop/internal/testsuite"
)

type Unittest struct {
	testsuite.LoggingSuite
}

func TestUnit(t *testing.T) { suite.Run(t, new(Unittest)) }

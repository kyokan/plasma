package cli

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/require"
	"os"
)

type CLITestSuite struct {
	suite.Suite
	tmpDir string
}

func (s *CLITestSuite) SetupSuite() {
	tmpDir, err := ioutil.TempDir("", "cli-tests")
	require.Nil(s.T(), err, "unexpected error")
	s.tmpDir = tmpDir
}

func (s *CLITestSuite) TearDownSuite() {
	err := os.RemoveAll(s.tmpDir)
	require.Nil(s.T(), err)
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}
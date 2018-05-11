package gitbase

import (
	"testing"

	"github.com/stretchr/testify/require"
	regression "gopkg.in/src-d/regression-core.v0"
)

func TestTest(t *testing.T) {
	require := require.New(t)

	config := regression.NewConfig()
	config.RepositoriesCache = "repo"
	config.BinaryCache = "binaries"
	config.Versions = []string{"remote:v0.11.0", "remote:master"}
	config.Complexity = 0
	config.Repeat = 1

	test, err := NewTest(config)
	require.NoError(err)

	err = test.Prepare()
	require.NoError(err)

	err = test.Run()
	require.NoError(err)

	test.GetResults()
}

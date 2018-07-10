package gitbase

import (
	"gopkg.in/src-d/regression-core.v0"
)

// NewToolGitbase creates a Tool with gitbase parameters filled.
func NewToolGitbase() regression.Tool {
	return regression.Tool{
		Name:        "gitbase",
		GitURL:      "https://github.com/src-d/gitbase",
		ProjectPath: "github.com/src-d/gitbase",
		BuildSteps: []regression.BuildStep{
			{
				Dir:     "",
				Command: "make",
				Args:    []string{"dependencies", "packages"},
			},
		},
	}
}

// NewGitbase returns a Binary struct for gitbase Tool.
func NewGitbase(
	config regression.Config,
	version string,
	releases *regression.Releases,
) *regression.Binary {
	return regression.NewBinary(config, NewToolGitbase(), version, releases)
}

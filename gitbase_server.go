package gitbase

import (
	"os"

	"github.com/src-d/regression-core"
)

// Server wraps a gitbase server instance.
type Server struct {
	*regression.Server
	binary    string
	repos     string
	indexPath string
}

// NewServer creates a new gitbase server struct.
func NewServer(binary, repos string) *Server {
	return &Server{
		Server: regression.NewServer(),
		binary: binary,
		repos:  repos,
	}
}

// URL returns the mysql URL to connect to gitbase server.
func (s *Server) URL() string {
	return "root@tcp(localhost)/"
}

// Start spawns a new gitbase server.
func (s *Server) Start() error {
	tmpDir, err := regression.CreateTempDir()
	if err != nil {
		return err
	}

	s.indexPath = tmpDir

	return s.Server.Start(
		s.binary,
		"server",
		"-g", s.repos,
		"-i", tmpDir,
	)
}

// Stops stops the gitbase server and deletes the index directory.
func (s *Server) Stop() (err error) {
	defer func() {
		rerr := os.RemoveAll(s.indexPath)
		if err == nil {
			err = rerr
		}
	}()

	err = s.Server.Stop()
	if err != nil {
		return err
	}

	return
}

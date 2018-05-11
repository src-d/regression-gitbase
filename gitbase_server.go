package gitbase

import (
	"gopkg.in/src-d/regression-core.v0"
)

// Server wraps a gitbase server instance.
type Server struct {
	*regression.Server
	binary string
	repos  string
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
	return s.Server.Start(s.binary, "server", "-g", s.repos)
}

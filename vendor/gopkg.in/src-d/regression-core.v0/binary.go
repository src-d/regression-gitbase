package regression

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/alcortesm/tgz"
	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"
)

var regRelease = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// ErrBinaryNotFound is returned when the executable is not found in
// the release tarball.
var ErrBinaryNotFound = errors.NewKind("binary not found in release tarball")

// Binary struct contains information and functionality to prepare and
// use a binary version.
type Binary struct {
	Version string
	Path    string

	releases *Releases
	config   Config
	tool     Tool
}

// NewBinary creates a new Binary structure.
func NewBinary(
	config Config,
	tool Tool,
	version string,
	releases *Releases,
) *Binary {
	return &Binary{
		Version:  version,
		releases: releases,
		config:   config,
		tool:     tool,
	}
}

// IsRelease checks if the version matches the format of a release, for
// example v0.12.1.
func (b *Binary) IsRelease() bool {
	return regRelease.MatchString(b.Version)
}

// Download prepares a binary version if it's still not in the
// binaries directory.
func (b *Binary) Download() error {
	switch {
	case IsRepo(b.Version):
		build, err := NewBuild(b.config, b.tool, b.Version)
		if err != nil {
			return err
		}

		binary, err := build.Build()
		if err != nil {
			return err
		}

		b.Path = binary
		return nil

	case b.Version == "latest":
		version, err := b.releases.Latest()
		if err != nil {
			return nil
		}

		b.Version = version

	case !b.IsRelease():
		b.Path = b.Version
		return nil
	}

	cacheName := b.cacheName()
	exist, err := fileExist(cacheName)
	if err != nil {
		return err
	}

	if exist {
		log.Debugf("Binary for %s already downloaded", b.Version)
		b.Path = cacheName
		return nil
	}

	log.Debugf("Dowloading version %s", b.Version)
	err = b.downloadRelease()
	if err != nil {
		log.Errorf(err, "Could not download version %s", b.Version)
		return err
	}

	b.Path = cacheName

	return nil
}

func (b *Binary) downloadRelease() error {
	tmpDir, err := CreateTempDir()
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	download := filepath.Join(tmpDir, "download.tar.gz")
	err = b.releases.Get(b.Version, b.tarName(), download)
	if err != nil {
		return err
	}

	path, err := tgz.Extract(download)
	if err != nil {
		return err
	}
	defer os.RemoveAll(path)

	binary := filepath.Join(path, b.dirName(), b.tool.Name)
	err = CopyFile(binary, b.cacheName(), 0755)

	return err
}

func (b *Binary) cacheName() string {
	binName := fmt.Sprintf("%s.%s", b.tool.Name, b.Version)
	return filepath.Join(b.config.BinaryCache, binName)
}

func (b *Binary) tarName() string {
	return fmt.Sprintf("%s_%s_%s_amd64.tar.gz",
		b.tool.Name,
		b.Version,
		b.config.OS,
	)
}

func (b *Binary) dirName() string {
	return b.tool.DirName(b.config.OS)
}

func fileExist(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

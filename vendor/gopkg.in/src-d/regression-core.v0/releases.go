package regression

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"gopkg.in/google/go-github.v15/github"
	"gopkg.in/src-d/go-errors.v1"
)

var releases *Releases

var (
	ErrVersionNotFound = errors.NewKind("Version '%s' not found")
	ErrAssetNotFound   = errors.NewKind(
		"Asset named '%s' not found in release '%s'")
)

type Releases struct {
	owner        string
	repo         string
	client       *github.Client
	repoReleases []*github.RepositoryRelease
}

func NewReleases(owner, repo, token string) *Releases {
	return &Releases{
		owner:  owner,
		repo:   repo,
		client: github.NewClient(oauthClient(token)),
	}
}

func oauthClient(token string) *http.Client {
	if token == "" {
		return nil
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	client := oauth2.NewClient(context.Background(), ts)

	return client
}

func (r *Releases) Get(version, asset, path string) error {
	err := r.getReleases()
	if err != nil {
		return err
	}

	for _, rel := range r.repoReleases {
		if rel.GetName() == version {
			for _, a := range rel.Assets {
				if a.GetName() == asset {
					return r.download(a.GetBrowserDownloadURL(), path)
				}
			}

			return ErrAssetNotFound.New(asset, version)
		}
	}

	return ErrVersionNotFound.New(version)
}

// Latest return the last version name from github releases
func (r *Releases) Latest() (string, error) {
	err := r.getReleases()
	if err != nil {
		return "", err
	}

	if len(r.repoReleases) < 1 {
		return "", ErrVersionNotFound.New("latest")
	}

	return r.repoReleases[0].GetName(), nil
}

func (r *Releases) getReleases() error {
	if r.repoReleases != nil {
		return nil
	}

	ctx := context.Background()
	rel, _, err := r.client.Repositories.ListReleases(ctx, r.owner, r.repo, nil)
	if err != nil {
		return err
	}

	r.repoReleases = rel
	return nil
}

func (r *Releases) download(url, path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	downloadPath := fmt.Sprintf("%s.download", path)
	exist, err := fileExist(downloadPath)
	if err != nil {
		return err
	}

	if exist {
		err = os.Remove(downloadPath)
		if err != nil {
			return err
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	err = os.Rename(downloadPath, path)
	return err
}

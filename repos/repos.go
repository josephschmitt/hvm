package repos

import (
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/josephschmitt/hvm/paths"
	log "github.com/sirupsen/logrus"
)

const DefaultRepository = "https://github.com/josephschmitt/hvm-packages.git"

type RepoLoader interface {
	Get() error
	Update() error
	Remove() error
}

type GitRepoLoader struct {
	Name     string
	Location string
	Ref      plumbing.ReferenceName
}

func NewGitRepoLoader(name string, url string) RepoLoader {
	loader := &GitRepoLoader{
		Location: url,
	}

	if loader.Location == "" {
		loader.Name = paths.PackageRepository
		loader.Location = DefaultRepository
		loader.Ref = "refs/heads/main"
	}

	return loader
}

func (g *GitRepoLoader) Get() error {
	log.Debugf("Get repo %s at %s\n", g.Name, g.Location)

	clonePath := filepath.Join(paths.AppPaths.ReposDirectory)

	w := log.New().WriterLevel(log.DebugLevel)
	defer w.Close()

	_, err := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL:           g.Location,
		Progress:      w,
		SingleBranch:  true,
		ReferenceName: g.Ref,
	})

	if err == git.ErrRepositoryAlreadyExists {
		return g.Update()
	}

	return err
}

func (g *GitRepoLoader) Update() error {
	log.Debugf("Update repo %s at %s\n", g.Name, g.Location)

	clonePath := filepath.Join(paths.AppPaths.ReposDirectory)

	w := log.New().WriterLevel(log.DebugLevel)
	defer w.Close()

	repo, err := git.PlainOpen(clonePath)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	alreadyUpToDate := false
	if err = worktree.Pull(&git.PullOptions{RemoteName: "origin"}); err != nil {
		if err.Error() != "already up-to-date" {
			return err
		}

		// If already up-to-date it throws an err, so just silence it here
		alreadyUpToDate = true
	}

	ref, err := repo.Head()
	if err != nil {
		return err
	}

	hash := ref.Hash()
	if alreadyUpToDate {
		log.Infof("Repository already up-to-date, at %s\n", hash)
	} else {
		log.Infof("Updated packages repository, now at %s\n", hash)
	}

	return nil
}

func (g *GitRepoLoader) Remove() error {
	return nil
}

type CurlRepoLoader struct {
	Location string
}

func (curl *CurlRepoLoader) Get() error {
	return nil
}

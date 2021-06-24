package paths

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
)

type Paths struct {
	GitRoot          string
	WorkingDirectory string
	HomeDirectory    string
	ConfigDirectory  string
	TempDirectory    string
}

func NewPaths() (*Paths, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine working directory")
	}
	return NewPathsFromDir(workDir)
}

func NewPathsFromDir(dir string) (*Paths, error) {
	gitRoot := FindDirGitRoot(dir)
	if gitRoot == "." {
		gitRoot = dir
	}

	u, err := user.Current()
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine current user")
	}

	tmpDir := os.Getenv("TMPDIR")
	if tmpDir == "" {
		tmpDir = "/tmp"
	}

	return &Paths{
		GitRoot:          gitRoot,
		WorkingDirectory: dir,
		HomeDirectory:    u.HomeDir,
		ConfigDirectory:  filepath.Join(u.HomeDir, ".hvm"),
		TempDirectory:    tmpDir,
	}, nil
}

func FindDirGitRoot(dir string) string {
	for dir != "/" {
		_, err := os.Stat(filepath.Join(dir, ".git"))
		if err != nil {
			if !os.IsNotExist(err) {
				return "."
			}
		} else {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return "."
}

func FindGitRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return FindDirGitRoot(dir)
}

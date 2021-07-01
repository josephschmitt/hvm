package paths

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

var AppPaths *Paths

const PackageRepository = "hvm-packages"
const PackageDownloads = "hvm-downloads"

type Paths struct {
	GitRoot          string
	WorkingDirectory string
	HomeDirectory    string
	ConfigDirectory  string
	TempDirectory    string
	ReposDirectory   string
	PkgsDirectory    string
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

	configDir := filepath.Join(u.HomeDir, ".hvm")

	return &Paths{
		GitRoot:          gitRoot,
		WorkingDirectory: dir,
		HomeDirectory:    u.HomeDir,
		ConfigDirectory:  configDir,
		TempDirectory:    filepath.Join(tmpDir, "hvm"),
		ReposDirectory:   filepath.Join(configDir, PackageRepository),
		PkgsDirectory:    filepath.Join(configDir, PackageDownloads),
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

func ConfigDirs(pths *Paths) []string {
	return []string{
		filepath.Join(pths.WorkingDirectory, ".hvm"),
		filepath.Join(pths.GitRoot, ".hvm"),
		filepath.Join(pths.HomeDirectory, ".hvm"),
	}
}

func ConfigFiles(pths *Paths) []string {
	var files []string
	for _, dir := range ConfigDirs(pths) {
		files = append(files, filepath.Join(dir, "config.hcl"))
	}

	return files
}

func (pths *Paths) ResolveDir(dir string) string {
	var homeDirRegexp = regexp.MustCompile(`^~|(?:\${?HOME}?)(/.*)?`)
	return homeDirRegexp.ReplaceAllString(dir, pths.HomeDirectory+"$1")
}

func init() {
	pths, err := NewPaths()
	if err != nil {
		panic(err)
	}

	AppPaths = pths
}

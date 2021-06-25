package context

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/hcl"

	"github.com/josephschmitt/hvm/paths"
	log "github.com/sirupsen/logrus"
)

type Context struct {
	Debug log.Level
}

type Config struct {
	Debug        string `kong:"default='warn',env='HVM_DEBUG'"`
	Repositories []string

	PackageConfig `hcl:"optional"`
}

func ConfigDirs(pths *paths.Paths) []string {
	return []string{
		filepath.Join(pths.WorkingDirectory, ".hvm"),
		filepath.Join(pths.GitRoot, ".hvm"),
		filepath.Join(pths.HomeDirectory, ".hvm"),
	}
}

func ConfigFiles(pths *paths.Paths) []string {
	var files []string
	for _, dir := range ConfigDirs(pths) {
		files = append(files, filepath.Join(dir, "config.hcl"))
	}

	return files
}

var xarch = map[string]string{
	"amd64":  "x64",
	"x86_64": "x64",
	"arm64":  "arm64",
}

func Platform() string {
	os := runtime.GOOS
	arch := xarch[runtime.GOARCH]

	return fmt.Sprintf("%s-%s", os, arch)
}

// PackageConfig is the reduced result of looking at all the app paths and deciding the specific
// package we plan on pulling from the package repository
type PackageConfig struct {
	Packages []PackageBlock `hcl:"package,block,optional"`
}

type PackageBlock struct {
	Name string `hcl:"name,label"`
	PackageOptions
}

type PackageOptions struct {
	Version  string `hcl:"version,optional"`
	Platform string `hcl:"platform,optional"`

	Exec    string            `hcl:"exec,optional"`
	Bins    map[string]string `hcl:"bins,optional"`
	Source  string            `hcl:"source,optional"`
	Extract string            `hcl:"extract,optional"`

	OutputDir string
}

func (pkgOpt *PackageOptions) Resolve(name string, pths *paths.Paths) *PackageOptions {
	configFiles := ConfigFiles(pths)

	for _, confPath := range configFiles {
		hclFile, err := os.ReadFile(confPath)
		if err != nil {
			continue
		}

		config := &PackageConfig{}
		hcl.Unmarshal(hclFile, config)

		for _, localPkgOpt := range config.Packages {
			if localPkgOpt.Name != name {
				continue
			}

			log.Debugf(colour.Sprintf("Found config options for ^3%s^R at ^5%s^R:\n%+v\n", name, confPath, config))

			if pkgOpt.Version == "" {
				pkgOpt.Version = localPkgOpt.Version
			}

			if pkgOpt.Platform == "" {
				pkgOpt.Platform = localPkgOpt.Platform
			}
		}

		if pkgOpt.Platform == "" {
			pkgOpt.Platform = Platform()
		}
	}

	log.Debugf(colour.Sprintf("Resolved PackageOptions %+v\n", pkgOpt))

	return pkgOpt
}

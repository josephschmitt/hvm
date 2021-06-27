package pkgs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/hcl"
	"github.com/josephschmitt/hvm/context"
	"github.com/josephschmitt/hvm/paths"
	"github.com/josephschmitt/hvm/repos"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasttemplate"
)

// PackageManifest contains the parsed result of the .hcl config file for a package. It's used to
// determine how to download a specific package project
type PackageManifest struct {
	Name        string `hcl:"name"`
	Description string `hcl:"description,optional"`
	Test        string `hcl:"test,optional"`

	context.Package
}

func (man *PackageManifest) Resolve(
	name string,
	opt *context.Package,
	pths *paths.Paths,
) (*PackageManifest, error) {
	if _, err := os.Stat(pths.ReposDirectory); os.IsNotExist(err) {
		loader := repos.NewGitRepoLoader("", "")
		if err := loader.Get(); err != nil {
			return nil, err
		}
	}

	man.Version = opt.Version
	man.Platform = opt.Platform
	man.Exec = opt.Exec
	man.Bins = opt.Bins
	man.Source = opt.Source
	man.Extract = opt.Extract
	man.OutputDir = opt.OutputDir

	if man.OutputDir == "" {
		man.OutputDir = filepath.Join(pths.PkgsDirectory, name, opt.Version)
	}

	configFilePath := filepath.Join(pths.ReposDirectory, name+".hcl")
	data, err := os.ReadFile(configFilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf(colour.Sprintf("no hvm-package found named \"^2%s^R\" at ^6%s^R\n"+
			"Try updating your packages, or contributing a new one for \"^2%s^R\".",
			name, configFilePath, name))
	} else if err != nil {
		return nil, err
	}

	t := fasttemplate.New(string(data), "${", "}")
	s := t.ExecuteString(map[string]interface{}{
		"version":  opt.Version,
		"platform": opt.Platform,
		"output":   man.OutputDir,
	})

	hcl.Unmarshal([]byte(s), man)

	log.Debugf("PackageManifest %+v\n", man)

	return man, nil
}

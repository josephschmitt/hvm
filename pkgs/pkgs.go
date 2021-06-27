package pkgs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/hcl"
	"github.com/imdario/mergo"
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
	pkg *context.Package,
	pths *paths.Paths,
) (*PackageManifest, error) {
	loader := repos.NewGitRepoLoader("", "")
	if _, err := os.Stat(pths.ReposDirectory); os.IsNotExist(err) {
		if err := loader.Get(); err != nil {
			return nil, err
		}
	} else {
		if err := loader.Update(); err != nil {
			return nil, err
		}
	}

	if pkg == nil {
		pkg = context.NewPackage()
	}

	if man.OutputDir == "" {
		if pkg.OutputDir != "" {
			man.OutputDir = pkg.OutputDir
		} else {
			man.OutputDir = filepath.Join(pths.PkgsDirectory, name, pkg.Version)
		}
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

	man.Render(data, pkg)

	if err := mergo.Merge(&man.Package, pkg, mergo.WithOverride); err != nil {
		log.Error(err)
		return nil, err
	}

	// Need to re-run the template parsing after merging the configs
	data, err = hcl.Marshal(man)
	if err != nil {
		return nil, err
	}
	man.Render(data, pkg)

	log.Debugf("PackageManifest %+v\n", man)

	return man, nil
}

func (man *PackageManifest) Render(data []byte, pkg *context.Package) *PackageManifest {
	t := fasttemplate.New(string(data), "${", "}")
	s := t.ExecuteString(map[string]interface{}{
		"version":    pkg.Version,
		"platform":   pkg.Platform,
		"x-platform": context.XPlatform(pkg.Platform),
		"output":     man.OutputDir,
	})

	hcl.Unmarshal([]byte(s), man)

	return man
}

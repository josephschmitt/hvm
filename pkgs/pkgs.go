package pkgs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/hcl"
	"github.com/josephschmitt/hvm/context"
	"github.com/josephschmitt/hvm/paths"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasttemplate"
)

const PackageRepository = "hvm-packages"
const PackageDownloads = "hvm-downloads"

var xarch = map[string]string{
	"amd64":  "x64",
	"x86_64": "x64",
	"arm64":  "arm64",
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
	configFiles := context.ConfigFiles(pths)

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
			pkgOpt.Platform = GetPlatform()
		}
	}

	log.Debugf(colour.Sprintf("Resolved PackageOptions %+v\n", pkgOpt))

	return pkgOpt
}

// PackageManifest contains the parsed result of the .hcl config file for a package. It's used to
// determine how to download a specific package project
type PackageManifest struct {
	Name        string `hcl:"name"`
	Description string `hcl:"description"`
	Test        string `hcl:"test"`

	PackageOptions
}

func NewPackageManifest(name string, opt *PackageOptions, pths *paths.Paths) (*PackageManifest, error) {
	log.Debugf("NewPackageManifest %s %+v\n", name, opt)

	man := &PackageManifest{}
	man.Version = opt.Version
	man.Platform = opt.Platform
	man.Exec = opt.Exec
	man.Bins = opt.Bins
	man.Source = opt.Source
	man.Extract = opt.Extract
	man.OutputDir = opt.OutputDir

	if man.OutputDir == "" {
		man.OutputDir = filepath.Join(pths.ConfigDirectory, PackageDownloads, name, opt.Version)
	}

	configFilePath := filepath.Join(pths.ConfigDirectory, PackageRepository, name+".hcl")
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

func GetPlatform() string {
	os := runtime.GOOS
	arch := xarch[runtime.GOARCH]

	return fmt.Sprintf("%s-%s", os, arch)
}

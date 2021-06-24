package pkgs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/hcl"
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

// Package is the reduced result of looking at all the app paths and deciding the specific
// package we plan on pulling from the package repository
type Package struct {
	Version  string `hcl:"version"`
	Platform string `hcl:"platform,optional"`
}

func NewPackage(name string, pths *paths.Paths) *Package {
	log.Debugf("NewPackage %s %+v\n", name, pths)

	dirs := []string{
		pths.WorkingDirectory,
		pths.GitRoot,
		pths.ConfigDirectory,
	}

	pkg := &Package{}

	for _, dir := range dirs {
		confPath := filepath.Join(dir, name+".hcl")
		hclFile, err := os.ReadFile(confPath)
		if err != nil {
			continue
		}

		localPkg := &Package{}
		hcl.Unmarshal(hclFile, localPkg)

		if pkg.Version == "" {
			pkg.Version = localPkg.Version
		}

		if pkg.Platform == "" {
			pkg.Platform = localPkg.Platform
		}
	}

	if pkg.Platform == "" {
		pkg.Platform = GetPlatform()
	}

	log.Debugf("Package %+v\n", pkg)

	return pkg
}

// PackageConfig contains the parsed result of the .hcl config file for a package. It's used to
// determine how to download a specific package project
type PackageConfig struct {
	Name        string            `hcl:"name"`
	Description string            `hcl:"description"`
	Test        string            `hcl:"test"`
	Binaries    []string          `hcl:"binaries"`
	Source      string            `hcl:"source"`
	Extract     []string          `hcl:"extract"`
	Links       map[string]string `hcl:"links"`

	OutputDir string
}

func NewPackageConfig(name string, pkg *Package, pths *paths.Paths) (*PackageConfig, error) {
	log.Debugf("NewPackageConfig %s %+v\n", name, pkg)

	conf := &PackageConfig{
		OutputDir: filepath.Join(pths.ConfigDirectory, PackageDownloads, name, "16.0.0"),
	}

	pkgFilePath := filepath.Join(pths.ConfigDirectory, PackageRepository, name+".hcl")
	data, err := os.ReadFile(pkgFilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf(colour.Sprintf("no hvm-package found named \"^2%s^R\" at ^6%s^R\n"+
			"Try updating your packages, or contributing a new one for \"^2%s^R\".",
			name, pkgFilePath, name))
	} else if err != nil {
		return nil, err
	}

	t := fasttemplate.New(string(data), "${", "}")
	s := t.ExecuteString(map[string]interface{}{
		// TODO: Hard-coded for now until we read config files from disk
		"version":  pkg.Version,
		"platform": pkg.Platform,
		"output":   conf.OutputDir,
	})

	hcl.Unmarshal([]byte(s), conf)

	log.Debugf("PackageConfig %+v\n", conf)

	return conf, nil
}

func GetPlatform() string {
	os := runtime.GOOS
	arch := xarch[runtime.GOARCH]

	return fmt.Sprintf("%s-%s", os, arch)
}

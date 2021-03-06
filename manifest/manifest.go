package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/hcl"
	"github.com/blang/semver/v4"
	"github.com/imdario/mergo"
	"github.com/josephschmitt/hvm/paths"
	"github.com/josephschmitt/hvm/repos"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasttemplate"
)

type PackageManifestOptions struct {
	Version string            `hcl:"version,optional"`
	Exec    string            `hcl:"exec,optional"`
	Bins    map[string]string `hcl:"bins,optional"`
	Source  string            `hcl:"source,optional"`
	Extract string            `hcl:"extract,optional"`
	Test    string            `hcl:"test,optional"`
}

// PackageManifest contains the parsed result of the .hcl config file for a package. It's used to
// determine how to download a specific package project
type PackageManifest struct {
	Name string

	PackageManifestOptions
}

func NewPackageManfiest(
	name string,
	ctx *PackageManifestContext,
	overrides *PackageManifestOptions,
) (*PackageManifest, error) {
	man := &PackageManifest{Name: name}
	if err := man.UpdateRepos(); err != nil {
		return nil, err
	}

	conf := &PackageManifestConfig{Name: name}
	if err := conf.Parse(); err != nil {
		return nil, err
	}

	if err := conf.Merge(overrides, ctx); err != nil {
		return nil, err
	}

	if err := conf.Render(ctx); err != nil {
		return nil, err
	}

	if err := mergo.Merge(&man.PackageManifestOptions, conf.PackageManifestOptions, mergo.WithOverride); err != nil {
		return nil, err
	}

	// If no bins set, assume the bin is named after the package
	if man.Bins == nil || len(man.Bins) == 0 {
		man.Bins = make(map[string]string)
		man.Bins[conf.Name] = conf.Name
	}

	log.Debugf("PackageManifest %+v\n", man)

	return man, nil
}

func (man *PackageManifest) UpdateRepos() error {
	if _, err := os.Stat(paths.AppPaths.ReposDirectory); os.IsNotExist(err) {
		loader := repos.NewGitRepoLoader("", "")
		if err := loader.Get(); err != nil {
			return err
		}
	}

	return nil
}

type PackageManifestConfig struct {
	Name        string `hcl:"name"`
	Description string `hcl:"description,optional"`

	PackageManifestOptions
	Versions []PackageManifestVersionBlock `hcl:"with-version,block,optional"`
}

func NewPackageManfiestConfig(name string) (*PackageManifestConfig, error) {
	conf := &PackageManifestConfig{}

	manTmpl, err := conf.GetManifestTemplate(name)
	if err != nil {
		return nil, err
	}

	if err := hcl.Unmarshal(manTmpl, conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func (conf *PackageManifestConfig) Merge(
	overrides *PackageManifestOptions,
	ctx *PackageManifestContext,
) error {
	if ctx.Version != "" {
		conf.Version = ctx.Version
	}

	ctxVer, err := semver.Parse(conf.Version)
	if err != nil {
		return err
	}

	for _, version := range conf.Versions {
		expectedRange, err := semver.ParseRange(version.Version)
		if err != nil || expectedRange == nil || !expectedRange(ctxVer) {
			continue
		}

		if err := mergo.Merge(&conf.PackageManifestOptions, version.PackageManifestOptions, mergo.WithOverride); err != nil {
			log.Error(err)
			return err
		}
	}

	if overrides != nil {
		if err := mergo.Merge(&conf.PackageManifestOptions, overrides, mergo.WithOverride); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

func (conf *PackageManifestConfig) Parse() error {
	data, err := conf.GetManifestTemplate(conf.Name)
	if err != nil {
		return err
	}

	return hcl.Unmarshal(data, conf)
}

func (conf *PackageManifestConfig) Render(ctx *PackageManifestContext) error {
	// Use version declared in the manifest as a "default"
	if ctx.Version == "" {
		ctx.Version = conf.Version
	}

	data, err := hcl.Marshal(conf)
	if err != nil {
		return err
	}

	t := fasttemplate.New(string(data), "${", "}")
	s := t.ExecuteString(map[string]interface{}{
		"version":    ctx.Version,
		"platform":   ctx.Platform,
		"x-platform": ctx.XPlatform,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"output":     ctx.OutputDir,
	})

	return hcl.Unmarshal([]byte(s), conf)
}

func (*PackageManifestConfig) GetManifestTemplate(name string) ([]byte, error) {
	configFilePath := filepath.Join(paths.AppPaths.ReposDirectory, name+".hcl")
	data, err := os.ReadFile(configFilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf(colour.Sprintf("no hvm-package found named \"^2%s^R\" at ^6%s^R\n"+
			"Try updating your packages, or contributing a new one for \"^2%s^R\".",
			name, configFilePath, name))
	} else if err != nil {
		return nil, err
	}

	return data, nil
}

type PackageManifestVersionBlock struct {
	Version string `hcl:"version,label"`
	PackageManifestOptions
}

type PackageManifestContext struct {
	Version   string
	Platform  string
	XPlatform string
	OutputDir string
}

func NewManifestContext(name string, version string) *PackageManifestContext {
	platform := Platform()

	return &PackageManifestContext{
		Version:   version,
		Platform:  platform,
		XPlatform: XPlatform(platform),
		OutputDir: filepath.Join(paths.AppPaths.PkgsDirectory, name, version),
	}
}

var arch = map[string]string{
	"amd64": "x64",
	"arm64": "arm64",
}

func Platform() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, arch[runtime.GOARCH])
}

var xarch = map[string]string{
	"amd64": "x86_64",
	"arm64": "arm64",
}

func XPlatform(platform string) string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, xarch[runtime.GOARCH])
}

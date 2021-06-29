package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/hcl"
	"github.com/imdario/mergo"
	"github.com/josephschmitt/hvm/paths"
	"github.com/josephschmitt/hvm/repos"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasttemplate"
)

type PackageManifestOptions struct {
	Exec    string            `hcl:"exec,optional"`
	Bins    map[string]string `hcl:"bins,optional"`
	Source  string            `hcl:"source,optional"`
	Extract string            `hcl:"extract,optional"`
	Test    string            `hcl:"test,optional"`
}

// PackageManifest contains the parsed result of the .hcl config file for a package. It's used to
// determine how to download a specific package project
type PackageManifest struct {
	Name    string
	Version string

	PackageManifestOptions
}

func NewPackageManfiest(
	name string,
	ctx *PackageManifestContext,
	overrides *PackageManifestOptions,
	pths *paths.Paths,
) (*PackageManifest, error) {
	man := &PackageManifest{
		Name:    name,
		Version: ctx.Version,
	}
	if err := man.UpdateRepos(pths); err != nil {
		return nil, err
	}

	conf := &PackageManifestConfig{}

	manTmpl, err := conf.GetManifestTemplate(name, pths)
	if err != nil {
		return nil, err
	}

	if err := conf.Render(manTmpl, ctx); err != nil {
		return nil, err
	}

	if err := man.Merge(ctx, conf, overrides); err != nil {
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

func (man *PackageManifest) UpdateRepos(pths *paths.Paths) error {
	if _, err := os.Stat(pths.ReposDirectory); os.IsNotExist(err) {
		loader := repos.NewGitRepoLoader("", "")
		if err := loader.Get(); err != nil {
			return err
		}
	}

	return nil
}

func (man *PackageManifest) Merge(
	ctx *PackageManifestContext,
	conf *PackageManifestConfig,
	overrides *PackageManifestOptions,
) error {
	if err := mergo.Merge(&man.PackageManifestOptions, conf.PackageManifestOptions, mergo.WithOverride); err != nil {
		log.Error(err)
		return err
	}

	for _, version := range conf.Versions {
		// TODO: Match version ranges at some point instead of just exact versions
		if ctx.Version != version.Version {
			continue
		}

		if err := mergo.Merge(man, version, mergo.WithOverride); err != nil {
			log.Error(err)
			return err
		}
	}

	if overrides != nil {
		if err := mergo.Merge(&conf, overrides, mergo.WithOverride); err != nil {
			log.Error(err)
			return err
		}

		// Marshal newly merged manifest back into an hcl template and re-render with new merged data
		manTmpl, err := hcl.Marshal(conf)
		if err != nil {
			return err
		}

		conf.Render(manTmpl, ctx)
	}

	return nil
}

type PackageManifestConfig struct {
	Name        string `hcl:"name"`
	Description string `hcl:"description,optional"`

	PackageManifestOptions
	Versions []PackageManifestVersionBlock `hcl:"version,block,optional"`
}

func NewPackageManfiestConfig(name string, pths *paths.Paths) (*PackageManifestConfig, error) {
	conf := &PackageManifestConfig{}

	manTmpl, err := conf.GetManifestTemplate(name, pths)
	if err != nil {
		return nil, err
	}

	if err := hcl.Unmarshal(manTmpl, conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func (conf *PackageManifestConfig) Render(data []byte, ctx *PackageManifestContext) error {
	t := fasttemplate.New(string(data), "${", "}")
	s := t.ExecuteString(map[string]interface{}{
		"version":    ctx.Version,
		"platform":   ctx.Platform,
		"x-platform": ctx.XPlatform,
		"output":     ctx.OutputDir,
	})

	if err := hcl.Unmarshal([]byte(s), conf); err != nil {
		return err
	}

	return nil
}

func (*PackageManifestConfig) GetManifestTemplate(name string, pths *paths.Paths) ([]byte, error) {
	configFilePath := filepath.Join(pths.ReposDirectory, name+".hcl")
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

func NewManifestContext(name string, version string, pths *paths.Paths) *PackageManifestContext {
	platform := Platform()

	return &PackageManifestContext{
		Version:   version,
		Platform:  platform,
		XPlatform: XPlatform(platform),
		OutputDir: filepath.Join(pths.PkgsDirectory, name, version),
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

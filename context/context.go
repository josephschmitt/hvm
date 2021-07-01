package context

import (
	"os"
	"path/filepath"

	"github.com/josephschmitt/hvm/manifest"
	"github.com/kardianos/osext"

	"github.com/alecthomas/hcl"
	"github.com/imdario/mergo"

	"github.com/josephschmitt/hvm/paths"
	log "github.com/sirupsen/logrus"
)

const DefaultLogLevel = "info"

type Context struct {
	Debug   *log.Level
	Use     map[string]string
	LinkDir string

	Repositories []string
	Packages     map[string]*manifest.PackageManifestOptions
}

func NewContext(logLevel string) (*Context, error) {
	ctx := &Context{}
	if err := ctx.Synthesize(); err != nil {
		return nil, err
	}

	if ctx.Debug == nil {
		ctx.SetLogLevel(DefaultLogLevel)
	}

	return ctx, nil
}

func (ctx *Context) SetLogLevel(level string) (log.Level, error) {
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		return 0, err
	}

	log.SetLevel(logLevel)
	ctx.Debug = &logLevel

	return logLevel, nil
}

func (ctx *Context) Synthesize() error {
	if ctx.Packages == nil {
		ctx.Packages = make(map[string]*manifest.PackageManifestOptions)
	}

	configFiles := paths.AppPaths.ConfigFiles()

	for _, confPath := range configFiles {
		hclFile, err := os.ReadFile(confPath)
		if err != nil {
			continue
		}

		foundConfig := &Config{}
		hcl.Unmarshal(hclFile, foundConfig)
		if err := ctx.Merge(foundConfig); err != nil {
			return err
		}
	}

	if ctx.LinkDir == "" {
		binPath, err := osext.Executable()
		if err != nil {
			return err
		}

		// Default link dir to wherever the hvm binary lives
		ctx.LinkDir = paths.AppPaths.ResolveDir(filepath.Dir(binPath))
	}

	return nil
}

func (ctx *Context) Merge(config *Config) error {
	for _, pkgConf := range config.Packages {
		pkgOpt := pkgConf.GetPackage()

		pkg := ctx.Packages[pkgConf.Name]
		if pkg == nil {
			pkg = pkgOpt
		}

		if err := mergo.Merge(pkg, pkgOpt, mergo.WithOverride); err != nil {
			return err
		}

		ctx.Packages[pkgConf.Name] = pkg
	}

	// Merge non-package fields
	if ctx.Debug == nil {
		ctx.SetLogLevel(config.Debug)
	}

	if len(config.Use) > 0 {
		ctx.Use = config.Use
	}

	if ctx.LinkDir == "" {
		ctx.LinkDir = paths.AppPaths.ResolveDir(config.LinkDir)
	}

	return nil
}

func (ctx *Context) UseVersion(name string, version string) {
	if ctx.Use == nil {
		ctx.Use = make(map[string]string)
	}
	ctx.Use[name] = version
}

// Config is the result of unmarshalling a config.hcl file
type Config struct {
	Debug    string            `hcl:"debug,optional"`
	Use      map[string]string `hcl:"use,optional"`
	LinkDir  string            `hcl:"linkdir,optional"`
	Packages []PackageBlock    `hcl:"package,block,optional"`
}

type PackageBlock struct {
	Name string `hcl:"name,label"`
	manifest.PackageManifestOptions
}

func (b *PackageBlock) GetPackage() *manifest.PackageManifestOptions {
	pkg := &manifest.PackageManifestOptions{}
	if err := mergo.Merge(pkg, b.PackageManifestOptions, mergo.WithOverride); err != nil {
		return nil
	}

	return pkg
}

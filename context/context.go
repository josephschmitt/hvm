package context

import (
	"fmt"
	"os"
	"runtime"

	"github.com/alecthomas/hcl"
	"github.com/imdario/mergo"

	"github.com/josephschmitt/hvm/paths"
	log "github.com/sirupsen/logrus"
)

const DefaultLogLevel = "info"

type Context struct {
	Debug *log.Level

	Repositories []string
	Packages     map[string]*Package
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

func (context *Context) Synthesize() error {
	if context.Packages == nil {
		context.Packages = make(map[string]*Package)
	}

	configFiles := paths.ConfigFiles(paths.AppPaths)

	for _, confPath := range configFiles {
		hclFile, err := os.ReadFile(confPath)
		if err != nil {
			continue
		}

		foundConfig := &Config{}
		hcl.Unmarshal(hclFile, foundConfig)
		if err := context.Merge(foundConfig); err != nil {
			return err
		}
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

	return nil
}

// Config is the result of unmarshalling a config.hcl file
type Config struct {
	Debug    string         `hcl:"debug,optional"`
	Packages []PackageBlock `hcl:"package,block,optional"`
}

func (conf *Config) FindPackage(name string) *Package {
	if len(conf.Packages) == 0 {
		return nil
	}

	for _, pkg := range conf.Packages {
		if name == pkg.Name {
			return pkg.GetPackage()
		}
	}

	return nil
}

type PackageBlock struct {
	Name string `hcl:"name,label"`
	Package
}

func (b *PackageBlock) GetPackage() *Package {
	pkg := NewPackage()
	if err := mergo.Merge(pkg, b.Package, mergo.WithOverride); err != nil {
		return nil
	}

	return pkg
}

type Package struct {
	Version  string `hcl:"version,optional"`
	Platform string `hcl:"platform,optional"`

	Exec    string            `hcl:"exec,optional"`
	Bins    map[string]string `hcl:"bins,optional"`
	Source  string            `hcl:"source,optional"`
	Extract string            `hcl:"extract,optional"`

	OutputDir string
}

func NewPackage() *Package {
	return &Package{
		Platform: Platform(),
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

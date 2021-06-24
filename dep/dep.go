package dep

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/hcl"
	"github.com/josephschmitt/hvm/paths"
	"github.com/valyala/fasttemplate"
)

type Config struct {
	Version   string
	Platform  string
	OutputDir string
}

type Dependency struct {
	Name        string   `hcl:"name"`
	Description string   `hcl:"description"`
	Test        string   `hcl:"test"`
	Binaries    []string `hcl:"binaries"`
	Source      string   `hcl:"source"`
	Extract     []string `hcl:"extract"`
}

func GetDepConfig(name string, pths *paths.Paths) *Config {
	// TODO: Hard-coded for now until we read config files from disk
	return &Config{
		Version:   "16.0.0",
		Platform:  "darwin-arm64",
		OutputDir: filepath.Join(pths.ConfigDirectory, name, "16.0.0"),
	}
}

func ResolveDep(name string, config *Config, pths *paths.Paths) (*Dependency, error) {
	data, err := os.ReadFile(filepath.Join(pths.ConfigDirectory, "deps", name+".hcl"))
	if err != nil {
		return nil, err
	}

	t := fasttemplate.New(string(data), "${", "}")
	s := t.ExecuteString(map[string]interface{}{
		"version":  config.Version,
		"platform": config.Platform,
		"output":   config.OutputDir,
	})

	repo := &Dependency{}
	hcl.Unmarshal([]byte(s), repo)

	return repo, nil
}

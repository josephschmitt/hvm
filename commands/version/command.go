package version

import (
	_ "embed"
	"os"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/kong"
	"github.com/josephschmitt/hvm/manifest"
)

func printVersion() error {
	man, err := manifest.GetManifest()
	if err != nil {
		return err
	}

	colour.Printf("hvm ^4%s^R\n", man.Version)
	return nil
}

type VersionFlag bool

func (*VersionFlag) BeforeApply() error {
	printVersion()
	os.Exit(0)
	return nil
}

type VersionCmd struct{}

func (*VersionCmd) Run(ctx *kong.Context) error {
	return printVersion()
}

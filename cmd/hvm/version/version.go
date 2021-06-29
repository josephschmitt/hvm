package version

import (
	_ "embed"
	"os"

	"github.com/josephschmitt/hvm/context"

	"github.com/alecthomas/colour"
)

func printVersion() error {
	man, err := GetVersionManifest()
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

func (*VersionCmd) Run(ctx *context.Context) error {
	return printVersion()
}

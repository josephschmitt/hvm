package main

import (
	_ "embed"
	"os"

	"github.com/josephschmitt/hvm/cmd/hvm/link"
	"github.com/josephschmitt/hvm/cmd/hvm/repos"
	"github.com/josephschmitt/hvm/cmd/hvm/run"
	"github.com/josephschmitt/hvm/cmd/hvm/unlink"
	"github.com/josephschmitt/hvm/cmd/hvm/version"
	"github.com/josephschmitt/hvm/context"

	"github.com/alecthomas/kong"
	"github.com/posener/complete"
	"github.com/willabides/kongplete"
)

var hvm struct {
	Debug string `kong:"default='warn',env='HVM_DEBUG'"`

	Version            version.VersionFlag          `kong:"help='Show version information.'"`
	VersionCmd         version.VersionCmd           `kong:"cmd,name='version',help='Show version information.'"`
	InstallCompletions kongplete.InstallCompletions `kong:"cmd,help='Install shell completions'"`

	Link        link.LinkCmd         `kong:"cmd,help='Link a new hermetic dependency library'"`
	UnLink      unlink.UnLinkCmd     `kong:"cmd,aliases='unlink',help='Unlink an existing hermetic dependency library'"`
	Run         run.RunCmd           `kong:"cmd,help='Run a hermetic dependency'"`
	UpdateRepos repos.UpdateReposCmd `kong:"cmd,help='Updates the list of packages from the packages repositories'"`
}

func main() {
	parser := kong.Must(&hvm, kong.HelpOptions{
		Tree: true,
	})

	kongplete.Complete(parser,
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
	)

	kCtx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	ctx, err := context.NewContext(hvm.Debug)
	kCtx.FatalIfErrorf(err)

	err = kCtx.Run(ctx)
	kCtx.FatalIfErrorf(err)
}

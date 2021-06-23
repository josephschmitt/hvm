package main

import (
	_ "embed"
	"os"

	"github.com/josephschmitt/hvm/commands/add"
	"github.com/josephschmitt/hvm/commands/version"
	"github.com/josephschmitt/hvm/context"
	log "github.com/sirupsen/logrus"

	"github.com/alecthomas/kong"
	"github.com/posener/complete"
	"github.com/willabides/kongplete"
)

var hvm struct {
	Debug string `kong:"default='warn'"`

	Version            version.VersionFlag          `kong:"help='Show version information.'"`
	VersionCmd         version.VersionCmd           `kong:"cmd,name='version',help='Show version information.'"`
	InstallCompletions kongplete.InstallCompletions `kong:"cmd,help='install shell completions'"`

	Add add.AddCmd `kong:"cmd,help='Add a new hermetic dependency library'"`
}

func main() {
	parser := kong.Must(&hvm, kong.HelpOptions{
		Tree: true,
	})

	kongplete.Complete(parser,
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
	)

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	debugLevel, err := log.ParseLevel(hvm.Debug)
	if err == nil {
		log.SetLevel(debugLevel)
	}

	err = ctx.Run(&context.Context{Debug: debugLevel})
	ctx.FatalIfErrorf(err)
}

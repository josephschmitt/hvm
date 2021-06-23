package main

import (
	_ "embed"
	"encoding/json"
	"os"

	"github.com/alecthomas/colour"

	"github.com/alecthomas/kong"
	"github.com/posener/complete"
	"github.com/willabides/kongplete"
)

//go:embed manifest.json
var manifestFile []byte

type manifest struct {
	Version string `json:"version"`
}

type versionFlag bool

func (h *versionFlag) AfterApply() error {
	man := &manifest{}
	err := json.Unmarshal(manifestFile, man)
	if err != nil {
		return err
	}

	colour.Printf("hvm ^4%s^R", man.Version)
	return nil
}

var hvm struct {
	Version versionFlag
}

func main() {
	parser := kong.Must(&hvm, kong.HelpOptions{
		Tree: true,
	})

	kongplete.Complete(parser,
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
	)

	_, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
}
